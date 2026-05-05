variable "aws_region" {
  description = "AWS region for this lab."
  type        = string
  default     = "ap-south-1"
}

variable "cluster_name" {
  description = "Unique EKS cluster name."
  type        = string
  default     = "session21-devops"
}

variable "kubernetes_version" {
  description = "An EKS Kubernetes version supported in the chosen region."
  type        = string
  default     = "1.33"
}

variable "node_instance_type" {
  description = "EC2 instance type for the managed node group."
  type        = string
  default     = "t3.medium"
}
