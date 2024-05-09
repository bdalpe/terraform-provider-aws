// Code generated by internal/generate/servicepackages/main.go; DO NOT EDIT.

package waf

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	waf_sdkv1 "github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  DataSourceIPSet,
			TypeName: "aws_waf_ipset",
		},
		{
			Factory:  DataSourceRateBasedRule,
			TypeName: "aws_waf_rate_based_rule",
		},
		{
			Factory:  DataSourceRule,
			TypeName: "aws_waf_rule",
		},
		{
			Factory:  DataSourceSubscribedRuleGroup,
			TypeName: "aws_waf_subscribed_rule_group",
		},
		{
			Factory:  DataSourceWebACL,
			TypeName: "aws_waf_web_acl",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceByteMatchSet,
			TypeName: "aws_waf_byte_match_set",
		},
		{
			Factory:  ResourceGeoMatchSet,
			TypeName: "aws_waf_geo_match_set",
		},
		{
			Factory:  ResourceIPSet,
			TypeName: "aws_waf_ipset",
		},
		{
			Factory:  ResourceRateBasedRule,
			TypeName: "aws_waf_rate_based_rule",
			Name:     "Rate Based Rule",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceRegexMatchSet,
			TypeName: "aws_waf_regex_match_set",
		},
		{
			Factory:  ResourceRegexPatternSet,
			TypeName: "aws_waf_regex_pattern_set",
		},
		{
			Factory:  ResourceRule,
			TypeName: "aws_waf_rule",
			Name:     "Rule",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceRuleGroup,
			TypeName: "aws_waf_rule_group",
			Name:     "Rule Group",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceSizeConstraintSet,
			TypeName: "aws_waf_size_constraint_set",
		},
		{
			Factory:  ResourceSQLInjectionMatchSet,
			TypeName: "aws_waf_sql_injection_match_set",
		},
		{
			Factory:  ResourceWebACL,
			TypeName: "aws_waf_web_acl",
			Name:     "Web ACL",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceXSSMatchSet,
			TypeName: "aws_waf_xss_match_set",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.WAF
}

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*waf_sdkv1.WAF, error) {
	sess := config["session"].(*session_sdkv1.Session)

	return waf_sdkv1.New(sess.Copy(&aws_sdkv1.Config{Endpoint: aws_sdkv1.String(config[names.AttrEndpoint].(string))})), nil
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
