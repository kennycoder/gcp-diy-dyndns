terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.34.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "random_id" "default" {
  byte_length = 8
}

resource "google_storage_bucket" "default" {
  name                        = "${random_id.default.hex}-gcf-source" # Every bucket name must be globally unique
  location                    = "US"
  uniform_bucket_level_access = true
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = ["allUsers"]
  }
}

data "archive_file" "default" {
  type        = "zip"
  output_path = "/tmp/function-source.zip"
  source_dir  = "../"
}
resource "google_storage_bucket_object" "object" {
  name   = "function-source.zip"
  bucket = google_storage_bucket.default.name
  source = data.archive_file.default.output_path # Add path to the zipped function source code
}

resource "google_cloudfunctions2_function" "diydns_function" {
  name        = "diydns-function"
  location    = var.region
  description = "Custom Dynamic DNS Function"

  build_config {
    runtime     = "go122"
    entry_point = "handleHTTP" # Set the entry point
    source {
      storage_source {
        bucket = google_storage_bucket.default.name
        object = google_storage_bucket_object.object.name
      }
    }
  }

  service_config {
    max_instance_count = 1
    available_memory   = "256M"
    timeout_seconds    = 60
    environment_variables = {
        PROJECT_ID = var.project_id
        KEY = var.key
    }    
    service_account_email = google_service_account.cloud_dns_sa.email
  }
}

# Create a Service Account for Cloud DNS access
resource "google_service_account" "cloud_dns_sa" {
  account_id = "cloud-dns-access"
  display_name = "Cloud DNS Access Service Account"
}

resource "google_project_iam_member" "dns" {
  project = var.project_id
  role     = "roles/dns.admin"
  member   = "serviceAccount:${google_service_account.cloud_dns_sa.email}"
}

resource "google_cloud_run_service_iam_policy" "noauth" {
   location    = google_cloudfunctions2_function.diydns_function.location
   project     = google_cloudfunctions2_function.diydns_function.project
   service     = google_cloudfunctions2_function.diydns_function.name

   policy_data = data.google_iam_policy.noauth.policy_data
}

output "function_uri" {
  value = google_cloudfunctions2_function.diydns_function.service_config[0].uri
}

# Define variables for the environment variables
variable "region" {
  type = string
  description = "Region to use"
}

variable "project_id" {
  type = string
  description = "Your project id"
}

variable "key" {
  type = string
  description = "API access key"
}
