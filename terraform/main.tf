variable "aws_secret_access_key" {
  type        = string
}

variable "aws_access_key_id" {
  type        = string
}

variable "aws_default_region" {
  type        = string
}

variable "s3_bucket" {
  type        = string
}
variable "sqs_queue" {
  type        = string
}

variable "s3_endpoint" {
    type        = string
  default     = "http://s3.localhost.localstack.cloud:4566"
}

variable "sqs_endpoint" {
    type        = string
  default     = "http://localhost:4566"
}

provider "aws" {

  access_key                  = var.aws_access_key_id
  secret_key                  = var.aws_secret_access_key
  region                      = var.aws_default_region

  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    s3             = var.s3_endpoint
    sqs           =  var.sqs_endpoint
  }
}

resource "aws_s3_bucket" "reports-s3-bucket" {
  bucket = var.s3_bucket
}

resource "aws_sqs_queue" "reports-sqs-queue" {
  name = var.sqs_queue
  delay_seconds = 5
    max_message_size = 2048
    message_retention_seconds = 86400
    receive_wait_time_seconds = 10

}