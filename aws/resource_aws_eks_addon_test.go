package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_eks_addon", &resource.Sweeper{
		Name: "aws_eks_addon",
		F:    testSweepEksAddons,
	})
}

func testSweepEksAddons(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).eksconn
	ctx := context.TODO()
	var sweeperErrs *multierror.Error

	input := &eks.ListClustersInput{MaxResults: aws.Int64(100)}
	err = conn.ListClustersPagesWithContext(ctx, input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		for _, cluster := range page.Clusters {
			clusterName := aws.StringValue(cluster)
			input := &eks.ListAddonsInput{
				ClusterName: aws.String(clusterName),
			}
			err := conn.ListAddonsPagesWithContext(ctx, input, func(page *eks.ListAddonsOutput, lastPage bool) bool {
				for _, addon := range page.Addons {
					addonName := aws.StringValue(addon)
					log.Printf("[INFO] Deleting EKS Addon %s from Cluster %s", addonName, clusterName)
					input := &eks.DeleteAddonInput{
						AddonName:   aws.String(addonName),
						ClusterName: aws.String(clusterName),
					}

					_, err := conn.DeleteAddonWithContext(ctx, input)

					if err != nil && !isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting EKS Addon %s from Cluster %s: %w", addonName, clusterName, err))
						continue
					}

					if err := waitForDeleteEksAddonDeleteContext(ctx, conn, clusterName, addonName, 5*time.Minute); err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error waiting for EKS Addon %s deletion: %w", addonName, err))
						continue
					}
				}
				return true
			})
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EKS Addons for Cluster %s: %w", clusterName, err))
			}
		}

		return true
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EKS Addon sweep for %s: %s", region, err)
		return sweeperErrs // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving EKS Clusters: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSEksAddon_basic(t *testing.T) {
	var addon eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	clusterResourceName := "aws_eks_cluster.test"
	addonResourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddon_Required(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, addonResourceName, &addon),
					testAccMatchResourceAttrRegionalARN(addonResourceName, "arn", "eks", regexp.MustCompile(fmt.Sprintf("addon/%s/%s/.+$", rName, addonName))),
					resource.TestCheckResourceAttrPair(addonResourceName, "cluster_name", clusterResourceName, "name"),
					resource.TestCheckResourceAttr(addonResourceName, "addon_name", addonName),
					resource.TestCheckResourceAttr(addonResourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(addonResourceName, "status", eks.AddonStatusActive),
					resource.TestCheckResourceAttr(addonResourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      addonResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEksAddon_disappears(t *testing.T) {
	var addon eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddon_Required(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					testAccCheckAWSEksAddonDisappears(ctx, &addon),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEksAddon_disappears_Cluster(t *testing.T) {
	var addon eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddon_Required(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					testAccCheckAWSEksClusterDisappears(ctx, &addon),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEksAddons_AddonVersion(t *testing.T) {
	var addon1, addon2 eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	addonVersion1 := "v1.6.3-eksbuild.1"
	addonVersion2 := "v1.7.5-eksbuild.1"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigAddonVersion(rName, addonName, addonVersion1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "addon_version", addonVersion1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resolve_conflicts"},
			},
			{
				Config: testAccAWSEksAddonConfigAddonVersion(rName, addonName, addonVersion2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "addon_version", addonVersion2),
				),
			},
		},
	})
}

func TestAccAWSEksAddons_ResolveConflicts(t *testing.T) {
	var addon1, addon2 eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigResolveConflicts(rName, addonName, eks.ResolveConflictsNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", eks.ResolveConflictsNone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resolve_conflicts"},
			},
			{
				Config: testAccAWSEksAddonConfigResolveConflicts(rName, addonName, eks.ResolveConflictsOverwrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "resolve_conflicts", eks.ResolveConflictsOverwrite),
				),
			},
		},
	})
}

func TestAccAWSEksAddons_ServiceAccountRoleArn(t *testing.T) {
	var addon eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	serviceRoleResourceName := "aws_iam_role.test-service-role"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigServiceAccountRoleArn(rName, addonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon),
					resource.TestCheckResourceAttrPair(resourceName, "service_account_role_arn", serviceRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEksAddons_Tags(t *testing.T) {
	var addon1, addon2, addon3 eks.Addon
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_addon.test"
	addonName := "vpc-cni"
	ctx := context.TODO()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t); testAccPreCheckAWSEksAddon(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksAddonConfigTags1(rName, addonName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEksAddonConfigTags2(rName, addonName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEksAddonConfigTags1(rName, addonName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksAddonExists(ctx, resourceName, &addon3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSEksAddonExists(ctx context.Context, resourceName string, addon *eks.Addon) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no EKS Addon ID is set")
		}

		clusterName, addonName, err := resourceAwsEksAddonParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).eksconn
		output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
			ClusterName: aws.String(clusterName),
			AddonName:   aws.String(addonName),
		})
		if err != nil {
			return err
		}

		if output == nil || output.Addon == nil {
			return fmt.Errorf("EKS Addon (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Addon.AddonName) != addonName {
			return fmt.Errorf("EKS Addon (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Addon.ClusterName) != clusterName {
			return fmt.Errorf("EKS Addon (%s) not found", rs.Primary.ID)
		}

		*addon = *output.Addon

		return nil
	}
}

func testAccCheckAWSEksAddonDestroy(s *terraform.State) error {
	ctx := context.TODO()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_addon" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).eksconn

		// Handle eventual consistency
		err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
			output, err := conn.DescribeAddonWithContext(ctx, &eks.DescribeAddonInput{
				AddonName:   aws.String(rs.Primary.ID),
				ClusterName: aws.String(rs.Primary.Attributes["cluster_name"]),
			})

			if err != nil {
				if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
					return nil
				}
				return resource.NonRetryableError(err)
			}

			if output != nil && output.Addon != nil && aws.StringValue(output.Addon.AddonName) == rs.Primary.ID {
				return resource.RetryableError(fmt.Errorf("EKS Addon %s still exists", rs.Primary.ID))
			}

			return nil
		})

		return err
	}

	return nil
}

func testAccCheckAWSEksAddonDisappears(ctx context.Context, addon *eks.Addon) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).eksconn

		input := &eks.DeleteAddonInput{
			ClusterName: addon.ClusterName,
			AddonName:   addon.AddonName,
		}

		_, err := conn.DeleteAddonWithContext(ctx, input)

		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		return waitForDeleteEksAddonDeleteContext(ctx, conn, aws.StringValue(addon.ClusterName), aws.StringValue(addon.AddonName), 5*time.Minute)
	}
}

func testAccCheckAWSEksClusterDisappears(ctx context.Context, addon *eks.Addon) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).eksconn

		input := &eks.DeleteClusterInput{
			Name: addon.ClusterName,
		}

		_, err := conn.DeleteClusterWithContext(ctx, input)

		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		return waitForDeleteEksCluster(conn, aws.StringValue(addon.ClusterName), 30*time.Minute)
	}
}

func testAccPreCheckAWSEksAddon(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).eksconn

	input := &eks.DescribeAddonVersionsInput{}

	_, err := conn.DescribeAddonVersions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSEksAddonConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name                          = "terraform-testacc-eks-cluster-base"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = "terraform-testacc-eks-cluster-base"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName)
}

func testAccAWSEksAddon_Required(rName, addonName string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q
}
`, rName, addonName))
}

func testAccAWSEksAddonConfigAddonVersion(rName, addonName, addonVersion string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name      = aws_eks_cluster.test.name
  addon_name        = %[2]q
  addon_version     = %[3]q
  resolve_conflicts = "OVERWRITE"
}
`, rName, addonName, addonVersion))
}

func testAccAWSEksAddonConfigResolveConflicts(rName, addonName, resolveConflicts string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name      = aws_eks_cluster.test.name
  addon_name        = %[2]q
  resolve_conflicts = %[3]q
}
`, rName, addonName, resolveConflicts))
}

func testAccAWSEksAddonConfigServiceAccountRoleArn(rName, addonName string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_iam_role" "test-service-role" {
  name               = "test-service-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_eks_addon" "test" {
  cluster_name             = aws_eks_cluster.test.name
  addon_name               = %[2]q
  service_account_role_arn = aws_iam_role.test-service-role.arn
}
`, rName, addonName))
}

func testAccAWSEksAddonConfigTags1(rName, addonName, tagKey1, tagValue1 string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, addonName, tagKey1, tagValue1))
}

func testAccAWSEksAddonConfigTags2(rName, addonName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(testAccAWSEksAddonConfig_Base(rName), fmt.Sprintf(`
resource "aws_eks_addon" "test" {
  cluster_name = aws_eks_cluster.test.name
  addon_name   = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, addonName, tagKey1, tagValue1, tagKey2, tagValue2))
}
