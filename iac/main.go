package main

import (
	"fmt"
	"os"

	"net/url"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/iam"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/lambda"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/logs"
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

var (
	commonTags = aws.TagArray{
		aws.TagArgs{
			Key:   pulumi.String("managedBy"),
			Value: pulumi.String("Pulumi"),
		},
	}
	err error

	lambdaImageTag = "45b5825"
	bucketOacId    = "E1UT8NIVK58ZX6"
)

func newDnsRecord(ctx *pulumi.Context, name string, dns string, zoneId pulumi.StringOutput, domain pulumi.StringOutput, recordType route53.RecordType, provider *aws.Provider) error {
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
	}, pulumi.DeleteBeforeReplace(true), pulumi.Provider(provider))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		config := config.New(ctx, "")
		accountId := config.RequireSecret("account-id")

		// for Route53
		managementProvider, err := aws.NewProvider(ctx, "management-provider", &aws.ProviderArgs{
			Profile: pulumi.String("default"),
			Region:  pulumi.String("us-east-1"),
		})
		if err != nil {
			return err
		}

		prodProvider, err := aws.NewProvider(ctx, "prod-provider", &aws.ProviderArgs{
			Profile: pulumi.String("prod"),
			Region:  pulumi.String("us-east-2"),
		})
		if err != nil {
			return err
		}

		// Lambda function

		logGroup, err := logs.NewLogGroup(ctx, "log-group", &logs.LogGroupArgs{
			LogGroupName:    pulumi.String("/aws/lambda/ipv6test"),
			RetentionInDays: pulumi.Int(7),
			Tags:            commonTags,
		}, pulumi.Provider(prodProvider))
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
		}, pulumi.Provider(prodProvider))
		if err != nil {
			return err
		}

		function, err := lambda.NewFunction(ctx, "function", &lambda.FunctionArgs{
			Architectures: &lambda.FunctionArchitecturesItemArray{
				lambda.FunctionArchitecturesItemArm64,
			},
			Code: lambda.FunctionCodeArgs{
				ImageUri: pulumi.Sprintf("%v.dkr.ecr.us-east-2.amazonaws.com/lambda/ipv6test:%s", accountId, lambdaImageTag), // placeholder; build in GitHub to deploy
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
		}, pulumi.Provider(prodProvider))
		if err != nil {
			return err
		}

		version, err := lambda.NewVersion(ctx, "version", &lambda.VersionArgs{
			FunctionName: function.FunctionName.Elem().ToStringOutput(),
		}, pulumi.Provider(prodProvider))
		if err != nil {
			return err
		}

		alias, err := lambda.NewAlias(ctx, "alias", &lambda.AliasArgs{
			FunctionVersion: version.Version,
			FunctionName:    function.FunctionName.Elem().ToStringOutput(),
			Name:            pulumi.String("main"),
			Description:     pulumi.String("Live site from main branch"),
		}, pulumi.Provider(prodProvider))
		if err != nil {
			return err
		}

		functionUrl, err := lambda.NewUrl(ctx, "function-url", &lambda.UrlArgs{
			AuthType:          lambda.UrlAuthTypeAwsIam,
			Qualifier:         alias.Name,
			TargetFunctionArn: function.Arn,
		}, pulumi.Provider(prodProvider))
		if err != nil {
			return err
		}

		// CloudFront function

		functionCode, err := os.ReadFile("cf-function/true-client-ip.js")
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
		}, pulumi.Provider(prodProvider))
		if err != nil {
			return err
		}

		// CloudFront access controls

		oacLambda, err := cloudfront.NewOriginAccessControl(ctx, "lambda", &cloudfront.OriginAccessControlArgs{
			OriginAccessControlConfig: &cloudfront.OriginAccessControlConfigArgs{
				Description:                   pulumi.String("Access to Lambda function"),
				Name:                          pulumi.String("lambda-access-ipv6test"),
				SigningBehavior:               pulumi.String("always"),
				SigningProtocol:               pulumi.String("sigv4"),
				OriginAccessControlOriginType: pulumi.String("lambda"),
			},
		}, pulumi.Provider(prodProvider))
		if err != nil {
			return err
		}

		// Response headers policies

		cacheOneDayPolicy, err := cloudfront.NewResponseHeadersPolicy(ctx, "one-day", &cloudfront.ResponseHeadersPolicyArgs{
			ResponseHeadersPolicyConfig: &cloudfront.ResponseHeadersPolicyConfigArgs{
				Comment: pulumi.String("Cache contents for one day"),
				CustomHeadersConfig: &cloudfront.ResponseHeadersPolicyCustomHeadersConfigArgs{
					Items: &cloudfront.ResponseHeadersPolicyCustomHeaderArray{
						cloudfront.ResponseHeadersPolicyCustomHeaderArgs{
							Header:   pulumi.String("Cache-Control"),
							Value:    pulumi.String("86400"),
							Override: pulumi.Bool(false),
						},
					},
				},
				Name: pulumi.String("cache-one-day"),
			},
		}, pulumi.Provider(prodProvider))
		if err != nil {
			return err
		}

		cacheOneWeekPolicy, err := cloudfront.NewResponseHeadersPolicy(ctx, "one-week", &cloudfront.ResponseHeadersPolicyArgs{
			ResponseHeadersPolicyConfig: &cloudfront.ResponseHeadersPolicyConfigArgs{
				Comment: pulumi.String("Cache contents for one week"),
				CustomHeadersConfig: &cloudfront.ResponseHeadersPolicyCustomHeadersConfigArgs{
					Items: &cloudfront.ResponseHeadersPolicyCustomHeaderArray{
						cloudfront.ResponseHeadersPolicyCustomHeaderArgs{
							Header:   pulumi.String("Cache-Control"),
							Value:    pulumi.String("604800"),
							Override: pulumi.Bool(false),
						},
					},
				},
				Name: pulumi.String("cache-one-week"),
			},
		}, pulumi.Provider(prodProvider))
		if err != nil {
			return err
		}

		cacheLambdaHeadersPolicy, err := cloudfront.NewResponseHeadersPolicy(ctx, "cache-lambda-index", &cloudfront.ResponseHeadersPolicyArgs{
			ResponseHeadersPolicyConfig: &cloudfront.ResponseHeadersPolicyConfigArgs{
				Comment: pulumi.String("Policy for uncached lambda function URLs"),
				CustomHeadersConfig: &cloudfront.ResponseHeadersPolicyCustomHeadersConfigArgs{
					Items: &cloudfront.ResponseHeadersPolicyCustomHeaderArray{
						cloudfront.ResponseHeadersPolicyCustomHeaderArgs{
							Header:   pulumi.String("Cache-Control"),
							Value:    pulumi.String("max-age=0, must-revalidate, private"),
							Override: pulumi.Bool(false),
						},
						cloudfront.ResponseHeadersPolicyCustomHeaderArgs{
							Header:   pulumi.String("Permissions-Policy"),
							Value:    pulumi.String("geolocation=()"),
							Override: pulumi.Bool(false),
						},
					},
				},
				Name: pulumi.String("lambda-index-page"),
				SecurityHeadersConfig: &cloudfront.ResponseHeadersPolicySecurityHeadersConfigArgs{
					ContentSecurityPolicy: &cloudfront.ResponseHeadersPolicyContentSecurityPolicyArgs{
						ContentSecurityPolicy: pulumi.String("script-src 'self'; frame-ancestors 'none'"),
						Override:              pulumi.Bool(false),
					},
					ContentTypeOptions: &cloudfront.ResponseHeadersPolicyContentTypeOptionsArgs{
						Override: pulumi.Bool(false),
					},
					ReferrerPolicy: &cloudfront.ResponseHeadersPolicyReferrerPolicyArgs{
						ReferrerPolicy: pulumi.String("no-referrer"),
						Override:       pulumi.Bool(false),
					},
					StrictTransportSecurity: &cloudfront.ResponseHeadersPolicyStrictTransportSecurityArgs{
						AccessControlMaxAgeSec: pulumi.Int(63072000),
						IncludeSubdomains:      pulumi.Bool(true),
						Override:               pulumi.Bool(false),
					},
					XssProtection: &cloudfront.ResponseHeadersPolicyXssProtectionArgs{
						ModeBlock:  pulumi.Bool(true),
						Override:   pulumi.Bool(false),
						Protection: pulumi.Bool(true),
					},
				},
			},
		}, pulumi.Provider(prodProvider))
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

		for _, distributionConfig := range distributionConfigs {
			distribution, err := cloudfront.NewDistribution(ctx, distributionConfig.dns, &cloudfront.DistributionArgs{
				DistributionConfig: &cloudfront.DistributionConfigArgs{
					Aliases: pulumi.ToStringArray(
						[]string{
							distributionConfig.dns,
						},
					),
					CacheBehaviors: cloudfront.DistributionCacheBehaviorArray{
						&cloudfront.DistributionCacheBehaviorArgs{
							AllowedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
								pulumi.String("OPTIONS"),
								pulumi.String("POST"),
								pulumi.String("PUT"),
								pulumi.String("PATCH"),
								pulumi.String("DELETE"),
							},
							CachedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
							},
							CachePolicyId:         pulumi.String(managedCachingDisabledPolicyId),
							OriginRequestPolicyId: pulumi.String(managedOriginRequestPolicyUserAgentRefererHeaders),
							ViewerProtocolPolicy:  pulumi.String("https-only"),
							PathPattern:           pulumi.String("/api/event"),
							TargetOriginId:        pulumi.String("plausible.io"),
							Compress:              pulumi.Bool(true),
						},
						&cloudfront.DistributionCacheBehaviorArgs{
							AllowedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
							},
							CachedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
							},
							CachePolicyId:        pulumi.String(managedCachingOptimizedPolicyId),
							ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
							PathPattern:          pulumi.String("/js/script.js"),
							TargetOriginId:       pulumi.String("plausible.io"),
							Compress:             pulumi.Bool(true),
						},
						&cloudfront.DistributionCacheBehaviorArgs{
							AllowedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
							},
							CachedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
							},
							CachePolicyId:           pulumi.String(managedCachingOptimizedPolicyId),
							ViewerProtocolPolicy:    pulumi.String("redirect-to-https"),
							PathPattern:             pulumi.String("/assets/*"),
							TargetOriginId:          pulumi.String("S3"),
							Compress:                pulumi.Bool(true),
							ResponseHeadersPolicyId: cacheOneDayPolicy.ID(),
						},
						&cloudfront.DistributionCacheBehaviorArgs{
							AllowedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
							},
							CachedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
							},
							CachePolicyId:           pulumi.String(managedCachingOptimizedPolicyId),
							ViewerProtocolPolicy:    pulumi.String("redirect-to-https"),
							PathPattern:             pulumi.String("/favicon.ico"),
							TargetOriginId:          pulumi.String("S3"),
							Compress:                pulumi.Bool(true),
							ResponseHeadersPolicyId: cacheOneDayPolicy.ID(),
						},
						&cloudfront.DistributionCacheBehaviorArgs{
							AllowedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
							},
							CachedMethods: pulumi.StringArray{
								pulumi.String("HEAD"),
								pulumi.String("GET"),
							},
							CachePolicyId:           pulumi.String(managedCachingOptimizedPolicyId),
							ViewerProtocolPolicy:    pulumi.String("redirect-to-https"),
							PathPattern:             pulumi.String("/robots.txt"),
							TargetOriginId:          pulumi.String("S3"),
							Compress:                pulumi.Bool(true),
							ResponseHeadersPolicyId: cacheOneWeekPolicy.ID(),
						},
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
						CachePolicyId:           pulumi.String(managedCachingDisabledPolicyId),
						Compress:                pulumi.Bool(true),
						OriginRequestPolicyId:   pulumi.String(managedOriginRequestPolicyAllViewerExceptHostHeader),
						ViewerProtocolPolicy:    pulumi.String("allow-all"),
						TargetOriginId:          pulumi.String("lambda"),
						ResponseHeadersPolicyId: cacheLambdaHeadersPolicy.ID(),
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
						&cloudfront.DistributionOriginArgs{
							ConnectionAttempts: pulumi.Int(3),
							ConnectionTimeout:  pulumi.Int(10),
							DomainName:         pulumi.String("plausible.io"),
							Id:                 pulumi.String("plausible.io"),
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
						&cloudfront.DistributionOriginArgs{
							ConnectionAttempts:    pulumi.Int(3),
							ConnectionTimeout:     pulumi.Int(10),
							DomainName:            config.RequireSecret("bucket-regional-domain-name"),
							OriginPath:            pulumi.String("/ipv6test.app"),
							Id:                    pulumi.String("S3"),
							OriginAccessControlId: pulumi.String(bucketOacId),
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
					WebAclId: config.RequireSecret("web-acl-id"),
					ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
						AcmCertificateArn:      distributionConfig.certArn,
						MinimumProtocolVersion: pulumi.String("TLSv1.2_2021"),
						SslSupportMethod:       pulumi.String("sni-only"),
					},
				},
				Tags: commonTags,
			}, pulumi.Provider(prodProvider))
			if err != nil {
				return err
			}

			_, err = lambda.NewPermission(ctx, fmt.Sprintf("%v-invokeUrl", distributionConfig.dns), &lambda.PermissionArgs{
				Action:       pulumi.String("lambda:InvokeFunctionUrl"),
				FunctionName: pulumi.Sprintf("%v:%v", function.Arn, alias.Name),
				Principal:    pulumi.String("cloudfront.amazonaws.com"),
				SourceArn:    pulumi.Sprintf("arn:aws:cloudfront::%v:distribution/%v", accountId, distribution.ID()),
			}, pulumi.Provider(prodProvider))
			if err != nil {
				return err
			}

			_, err = lambda.NewPermission(ctx, fmt.Sprintf("%v-invoke", distributionConfig.dns), &lambda.PermissionArgs{
				Action:       pulumi.String("lambda:InvokeFunction"),
				FunctionName: pulumi.Sprintf("%v:%v", function.Arn, alias.Name),
				Principal:    pulumi.String("cloudfront.amazonaws.com"),
				SourceArn:    pulumi.Sprintf("arn:aws:cloudfront::%v:distribution/%v", accountId, distribution.ID()),
			}, pulumi.Provider(prodProvider))
			if err != nil {
				return err
			}

			zoneId := config.RequireSecret("route53-zone-id")

			if distributionConfig.dns == "ipv6test.app" {
				err = newDnsRecord(ctx, "ipv6test-a", "ipv6test.app", zoneId, distribution.DomainName, route53.RecordTypeA, managementProvider)
				if err != nil {
					return err
				}

				err = newDnsRecord(ctx, "ipv6test-aaaa", "ipv6test.app", zoneId, distribution.DomainName, route53.RecordTypeAAAA, managementProvider)
				if err != nil {
					return err
				}
			}

			if distributionConfig.dns == "v6.ipv6test.app" {
				err = newDnsRecord(ctx, "v6-ipv6test-aaaa", "v6.ipv6test.app", zoneId, distribution.DomainName, route53.RecordTypeAAAA, managementProvider)
				if err != nil {
					return err
				}
			}

			if distributionConfig.dns == "v4.ipv6test.app" {
				err = newDnsRecord(ctx, "v4-ipv6test-a", "v4.ipv6test.app", zoneId, distribution.DomainName, route53.RecordTypeA, managementProvider)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}
