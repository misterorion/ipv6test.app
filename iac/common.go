package main

import (
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/cloudfront"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var plausibleOrigin = &cloudfront.DistributionOriginArgs{
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
}

var plausibleApiCacheBehavior = &cloudfront.DistributionCacheBehaviorArgs{
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
}

var plausibleScriptCacheBehavior = &cloudfront.DistributionCacheBehaviorArgs{
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
}

var robotsCacheBehavior = &cloudfront.DistributionCacheBehaviorArgs{
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
	PathPattern:          pulumi.String("/robots.txt"),
	TargetOriginId:       pulumi.String("S3"),
	Compress:             pulumi.Bool(true),
}

var assetsCacheBehavior = &cloudfront.DistributionCacheBehaviorArgs{
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
	PathPattern:          pulumi.String("/assets/*"),
	TargetOriginId:       pulumi.String("S3"),
	Compress:             pulumi.Bool(true),
}

var faviconCacheBehavior = &cloudfront.DistributionCacheBehaviorArgs{
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
	PathPattern:          pulumi.String("/favicon.ico"),
	TargetOriginId:       pulumi.String("S3"),
	Compress:             pulumi.Bool(true),
}

var commonTags = aws.TagArray{
	aws.TagArgs{
		Key:   pulumi.String("managedBy"),
		Value: pulumi.String("Pulumi"),
	},
}
