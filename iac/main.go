package main

import (
	"os"

	"net/url"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecr"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/iam"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/lambda"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/logs"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/s3"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	cloudFrontDefaultZoneId                             = "Z2FDTNDATAQYW2"
	managedCachingDisabledPolicyId                      = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad"
	managedCachingOptimizedPolicyId                     = "658327ea-f89d-4fab-a63d-7e88639e58f6"
	managedOriginRequestPolicyUserAgentRefererHeaders   = "acba4595-bd28-49b8-b9fe-13317c0390fa"
	managedOriginRequestPolicyAllViewerExceptHostHeader = "b689b0a8-53d0-40ab-baf2-68738e2966ac"
)

var err error

func newDnsRecord(ctx *pulumi.Context, name string, dns string, zoneId pulumi.StringOutput, domain pulumi.StringOutput, recordType route53.RecordType) error {
	_, err = route53.NewRecord(ctx, name, &route53.RecordArgs{
		ZoneId: zoneId,
		Name:   pulumi.String(dns),
		Type:   pulumi.String(recordType),
		Aliases: route53.RecordAliasArray{
			route53.RecordAliasArgs{
				EvaluateTargetHealth: pulumi.Bool(false),
				Name:                 domain,
				ZoneId:               pulumi.String(cloudFrontDefaultZoneId),
			},
		},
	}, pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		config := config.New(ctx, "")
		accountId := config.RequireSecret("account-id")
		bucketName := config.RequireSecret("bucket-name")

		// S3 Bucket for Assets

		bucket, err := s3.NewBucket(ctx, "bucket", &s3.BucketArgs{
			BucketName: bucketName,
		})
		if err != nil {
			return err
		}

		// Lambda

		_, err = ecr.NewRepository(ctx, "ecr-repo", &ecr.RepositoryArgs{
			EmptyOnDelete: pulumi.Bool(true),
			ImageScanningConfiguration: ecr.RepositoryImageScanningConfigurationArgs{
				ScanOnPush: pulumi.Bool(true),
			},
			ImageTagMutability: ecr.RepositoryImageTagMutabilityImmutable,
			LifecyclePolicy: &ecr.RepositoryLifecyclePolicyArgs{
				LifecyclePolicyText: pulumi.String(`{
					"rules": [
						{
							"rulePriority": 1,
							"description": "Expire old images",
							"selection": {
								"tagStatus": "any",
								"countType": "imageCountMoreThan",
								"countNumber": 2
							},
							"action": {
								"type": "expire"
							}
						}
					]
				}`),
			},
			RepositoryName: pulumi.String("lambda/ipv6test"),
			Tags:           commonTags,
		})
		if err != nil {
			return err
		}

		logGroup, err := logs.NewLogGroup(ctx, "log-group", &logs.LogGroupArgs{
			LogGroupName:    pulumi.String("/aws/lambda/ipv6test"),
			RetentionInDays: pulumi.Int(1),
			Tags:            commonTags,
		})
		if err != nil {
			return err
		}

		role, err := iam.NewRole(ctx, "lambda-iam-role", &iam.RoleArgs{
			AssumeRolePolicyDocument: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
				  {
					"Effect": "Allow",
					"Principal": {
					  "Service": "lambda.amazonaws.com"
					},
					"Action": "sts:AssumeRole"
				  }
				]
			  }`),
			ManagedPolicyArns: pulumi.StringArray{
				pulumi.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
			},
			RoleName: pulumi.String("lambda-ipv6test-role"),
			Tags:     commonTags,
		})
		if err != nil {
			return err
		}

		function, err := lambda.NewFunction(ctx, "function", &lambda.FunctionArgs{
			Architectures: &lambda.FunctionArchitecturesItemArray{
				lambda.FunctionArchitecturesItemArm64,
			},
			Code: lambda.FunctionCodeArgs{
				ImageUri: pulumi.Sprintf("%v.dkr.ecr.us-east-2.amazonaws.com/base:latest", accountId), // placeholder; build in GitHub to deploy
			},
			Description:  pulumi.String("Send static web page"),
			FunctionName: pulumi.String("ipv6test"),
			LoggingConfig: &lambda.FunctionLoggingConfigArgs{
				LogGroup: logGroup.LogGroupName,
			},
			PackageType:                  lambda.FunctionPackageTypeImage,
			ReservedConcurrentExecutions: pulumi.Int(5),
			Role:                         role.Arn,
			Tags:                         commonTags,
			Timeout:                      pulumi.Int(30),
		})
		if err != nil {
			return err
		}

		version, err := lambda.NewVersion(ctx, "version", &lambda.VersionArgs{
			FunctionName: function.FunctionName.Elem().ToStringOutput(),
		})
		if err != nil {
			return err
		}

		alias, err := lambda.NewAlias(ctx, "alias", &lambda.AliasArgs{
			FunctionVersion: version.Version,
			FunctionName:    function.FunctionName.Elem().ToStringOutput(),
			Name:            pulumi.String("main"),
			Description:     pulumi.String("Live site from main branch"),
		})
		if err != nil {
			return err
		}

		functionUrl, err := lambda.NewUrl(ctx, "function-url", &lambda.UrlArgs{
			AuthType:          lambda.UrlAuthTypeAwsIam,
			Qualifier:         alias.Name,
			TargetFunctionArn: function.Arn,
		})
		if err != nil {
			return err
		}

		// CloudFront function

		functionCode, err := os.ReadFile("function-code/true-client-ip.js")
		if err != nil {
			return err
		}

		trueClientIpFunction, err := cloudfront.NewFunction(ctx, "true-client-ip", &cloudfront.FunctionArgs{
			AutoPublish: pulumi.Bool(true),
			FunctionConfig: &cloudfront.FunctionConfigArgs{
				Comment: pulumi.String("Pass viewer IP address"),
				Runtime: pulumi.String("cloudfront-js-2.0"),
			},
			FunctionCode: pulumi.String(functionCode),
			Name:         pulumi.String("pass-true-client-ip"),
		})
		if err != nil {
			return err
		}

		// CloudFront access

		oacS3, err := cloudfront.NewOriginAccessControl(ctx, "s3", &cloudfront.OriginAccessControlArgs{
			OriginAccessControlConfig: &cloudfront.OriginAccessControlConfigArgs{
				Description:                   pulumi.String("Access to static assets"),
				Name:                          pulumi.String("s3-bucket-access"),
				SigningBehavior:               pulumi.String("always"),
				SigningProtocol:               pulumi.String("sigv4"),
				OriginAccessControlOriginType: pulumi.String("s3"),
			},
		})
		if err != nil {
			return err
		}

		oacLambda, err := cloudfront.NewOriginAccessControl(ctx, "lambda", &cloudfront.OriginAccessControlArgs{
			OriginAccessControlConfig: &cloudfront.OriginAccessControlConfigArgs{
				Description:                   pulumi.String("Access to Lambda function"),
				Name:                          pulumi.String("Lambda function access"),
				SigningBehavior:               pulumi.String("always"),
				SigningProtocol:               pulumi.String("sigv4"),
				OriginAccessControlOriginType: pulumi.String("lambda"),
			},
		})
		if err != nil {
			return err
		}

		// CloudFront distributions

		type distributionConfig struct {
			dns     string
			certArn pulumi.StringOutput
		}

		distributionConfigs := []distributionConfig{
			{
				dns:     "ipv6test.app",
				certArn: config.RequireSecret("certArn-ipv6test.app"),
			},
			{
				dns:     "v6.ipv6test.app",
				certArn: config.RequireSecret("certArn-v6.ipv6test.app"),
			},
			{
				dns:     "v4.ipv6test.app",
				certArn: config.RequireSecret("certArn-v4.ipv6test.app"),
			},
		}

		distributionIds := []pulumi.IDOutput{}

		for _, distributionConfig := range distributionConfigs {
			distribution, err := cloudfront.NewDistribution(ctx, distributionConfig.dns, &cloudfront.DistributionArgs{
				DistributionConfig: &cloudfront.DistributionConfigArgs{
					Aliases: pulumi.ToStringArray(
						[]string{
							distributionConfig.dns,
						},
					),
					CacheBehaviors: cloudfront.DistributionCacheBehaviorArray{
						plausibleApiCacheBehavior,
						plausibleScriptCacheBehavior,
						assetsCacheBehavior,
						faviconCacheBehavior,
						robotsCacheBehavior,
					},
					Comment: pulumi.String(distributionConfig.dns),
					DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
						AllowedMethods: pulumi.StringArray{
							pulumi.String("HEAD"),
							pulumi.String("GET"),
						},
						CachedMethods: pulumi.StringArray{
							pulumi.String("HEAD"),
							pulumi.String("GET"),
						},
						CachePolicyId:         pulumi.String(managedCachingDisabledPolicyId),
						Compress:              pulumi.Bool(true),
						OriginRequestPolicyId: pulumi.String(managedOriginRequestPolicyAllViewerExceptHostHeader),
						ViewerProtocolPolicy:  pulumi.String("allow-all"),
						TargetOriginId:        pulumi.String("lambda"),
						FunctionAssociations: cloudfront.DistributionFunctionAssociationArray{
							cloudfront.DistributionFunctionAssociationArgs{
								EventType:   pulumi.String("viewer-request"),
								FunctionArn: trueClientIpFunction.FunctionArn,
							},
						},
					},
					Enabled:     pulumi.Bool(true),
					HttpVersion: pulumi.String("http2and3"),
					Ipv6Enabled: pulumi.Bool(true),
					Origins: &cloudfront.DistributionOriginArray{
						plausibleOrigin,
						&cloudfront.DistributionOriginArgs{
							ConnectionAttempts:    pulumi.Int(3),
							ConnectionTimeout:     pulumi.Int(10),
							DomainName:            bucket.DomainName,
							Id:                    pulumi.String("S3"),
							OriginAccessControlId: oacS3.ID(),
							S3OriginConfig: cloudfront.DistributionS3OriginConfigArgs{
								OriginAccessIdentity: pulumi.String(""),
							},
						},
						&cloudfront.DistributionOriginArgs{
							ConnectionAttempts: pulumi.Int(3),
							ConnectionTimeout:  pulumi.Int(10),
							DomainName: functionUrl.FunctionUrl.ApplyT(func(u string) string {
								p, _ := url.Parse(u)
								return p.Hostname()
							}).(pulumi.StringOutput),
							Id: pulumi.String("lambda"),
							OriginCustomHeaders: cloudfront.DistributionOriginCustomHeaderArray{
								cloudfront.DistributionOriginCustomHeaderArgs{
									HeaderName:  pulumi.String("x-cdn-host"),
									HeaderValue: pulumi.String(distributionConfig.dns),
								},
							},
							OriginAccessControlId: oacLambda.ID(),
							CustomOriginConfig: &cloudfront.DistributionCustomOriginConfigArgs{
								HttpPort:               pulumi.Int(80),
								HttpsPort:              pulumi.Int(443),
								OriginKeepaliveTimeout: pulumi.Int(5),
								OriginProtocolPolicy:   pulumi.String("https-only"),
								OriginReadTimeout:      pulumi.Int(30),
								OriginSslProtocols: pulumi.StringArray{
									pulumi.String("TLSv1.2"),
								},
							},
						},
					},
					ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
						AcmCertificateArn:      distributionConfig.certArn,
						MinimumProtocolVersion: pulumi.String("TLSv1.2_2021"),
						SslSupportMethod:       pulumi.String("sni-only"),
					},
				},
				Tags: commonTags,
			})
			if err != nil {
				return err
			}

			distributionIds = append(distributionIds, distribution.ID())

			_, err = lambda.NewPermission(ctx, distributionConfig.dns, &lambda.PermissionArgs{
				Action:       pulumi.String("lambda:InvokeFunctionUrl"),
				FunctionName: pulumi.Sprintf("%v:%v", function.Arn, alias.Name),
				Principal:    pulumi.String("cloudfront.amazonaws.com"),
				SourceArn:    pulumi.Sprintf("arn:aws:cloudfront::%v:distribution/%v", accountId, distribution.ID()),
			})
			if err != nil {
				return err
			}

			zoneId := config.RequireSecret("route53-zone-id")
			if distributionConfig.dns == "ipv6test.app" {
				err = newDnsRecord(ctx, "ipv6test-a", "ipv6test.app", zoneId, distribution.DomainName, route53.RecordTypeA)
				if err != nil {
					return err
				}

				err = newDnsRecord(ctx, "ipv6test-aaaa", "ipv6test.app", zoneId, distribution.DomainName, route53.RecordTypeAAAA)
				if err != nil {
					return err
				}
			}

			if distributionConfig.dns == "v6.ipv6test.app" {
				err = newDnsRecord(ctx, "v6-ipv6test-aaaa", "v6.ipv6test.app", zoneId, distribution.DomainName, route53.RecordTypeAAAA)
				if err != nil {
					return err
				}
			}

			if distributionConfig.dns == "v4.ipv6test.app" {
				err = newDnsRecord(ctx, "v4-ipv6test-a", "v4.ipv6test.app", zoneId, distribution.DomainName, route53.RecordTypeA)
				if err != nil {
					return err
				}
			}
		}

		_, err = s3.NewBucketPolicy(ctx, "policy", &s3.BucketPolicyArgs{
			Bucket: bucketName,
			PolicyDocument: pulumi.Sprintf(`{
				"Version": "2008-10-17",
				"Id": "PolicyForCloudFrontPrivateContent",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {
							"Service": "cloudfront.amazonaws.com"
						},
						"Action": "s3:GetObject",
						"Resource": "arn:aws:s3:::%v/*",
						"Condition": {
							"StringLike": {
								"AWS:SourceArn": [
									"arn:aws:cloudfront::%v:distribution/%v",
									"arn:aws:cloudfront::%v:distribution/%v",
									"arn:aws:cloudfront::%v:distribution/%v"
								]
							}
						}
					}
				]
			}`, bucketName, accountId, distributionIds[0], accountId, distributionIds[1], accountId, distributionIds[2]),
		})
		if err != nil {
			return err
		}

		return nil
	})
}
