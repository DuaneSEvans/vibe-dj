import * as cdk from "aws-cdk-lib"
import { Construct } from "constructs"
import * as ec2 from "aws-cdk-lib/aws-ec2"
import * as ecs from "aws-cdk-lib/aws-ecs"
import * as elbv2 from "aws-cdk-lib/aws-elasticloadbalancingv2"
import * as ecr_assets from "aws-cdk-lib/aws-ecr-assets"
import * as path from "path"
import * as logs from "aws-cdk-lib/aws-logs"

export class InfraStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props)

    // Create a new VPC
    const vpc = new ec2.Vpc(this, "VibeDjVpc", {
      maxAzs: 2,
      subnetConfiguration: [
        {
          cidrMask: 24,
          name: "public-subnet",
          subnetType: ec2.SubnetType.PUBLIC,
        },
      ],
    })

    // Create an ECS cluster
    const cluster = new ecs.Cluster(this, "VibeDjCluster", {
      vpc,
    })

    // Build Docker image from /server and push to ECR
    const imageAsset = new ecr_assets.DockerImageAsset(
      this,
      "VibeDjImageAsset",
      {
        directory: path.join(__dirname, "../../"),
        exclude: ["**/cdk.out", "**/node_modules", ".git", ".env"],
      }
    )

    // Create a Fargate task definition
    const taskDefinition = new ecs.FargateTaskDefinition(
      this,
      "VibeDjTaskDef",
      {
        cpu: 256,
        memoryLimitMiB: 512,
      }
    )

    const container = taskDefinition.addContainer("VibeDjContainer", {
      image: ecs.ContainerImage.fromDockerImageAsset(imageAsset),
      portMappings: [
        {
          containerPort: 8080,
        },
      ],
      logging: ecs.LogDrivers.awsLogs({
        streamPrefix: "VibeDjService",
        logRetention: logs.RetentionDays.ONE_WEEK,
      }),
    })

    // Create a Fargate service
    const service = new ecs.FargateService(this, "VibeDjService", {
      cluster,
      taskDefinition,
      assignPublicIp: true,
      healthCheckGracePeriod: cdk.Duration.seconds(60),
    })

    // Create a load balancer
    const lb = new elbv2.ApplicationLoadBalancer(this, "VibeDjLb", {
      vpc,
      internetFacing: true,
    })

    const listener = lb.addListener("Listener", {
      port: 80,
    })

    listener.addTargets("ECS", {
      port: 8080,
      targets: [service],
      healthCheck: {
        path: "/",
        interval: cdk.Duration.seconds(30),
        timeout: cdk.Duration.seconds(5),
        healthyThresholdCount: 5,
      },
    })

    new cdk.CfnOutput(this, "LoadBalancerDns", {
      value: lb.loadBalancerDnsName,
    })
  }
}
