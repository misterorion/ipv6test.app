# ipv6test.app

An application I threw together to check if a client is IPv6-ready.

It consists of a single page web app deployed to AWS as a Lambda function fronted by 3 CloudFront distributions. The main distribution has both A and AAAA DNS entries. The other two distributions have only A or AAAA DNS entries.

The infrastructure-as-code is written in Go and deployed with Pulumi.

Live site: [https://ipv6test.app](https://ipv6test.app)
