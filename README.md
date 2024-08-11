# DIY Dynamic DNS on Google Cloud Functions

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Python](https://img.shields.io/badge/go-1.22-blue.svg)](https://www.go.dev/)
[![Terraform](https://img.shields.io/badge/terraform-%235835CC.svg)](https://www.terraform.io/)

**DIYDynDNS** is a simple api endpoint that listens for HTTPs requests, extracts the IP address of the caller (via X-Forwarded-For header) and updates the specified A entry to point to that IP address. 

Whole purpose of this mini-project is to overcome the dynamic ip allocation by most ISPs.


> :warning: Make sure that the API key is safely stored, otherwise anyone would be able to update your DNS records!

## Getting Started

### Prerequisites

* Google Cloud project/account
* [gcloud cli](https://cloud.google.com/sdk/docs/install)
* [Terraform](https://developer.hashicorp.com/terraform/install)

### Deploying

Before anything, make sure you are authenticated to Google Cloud Platform:

1. Run `gcloud auth login` and `gcloud auth application-default login`.
2. Enable necessary APIS: 
<pre>
gcloud services enable cloudbuild.googleapis.com \
cloudfunctions.googleapis.com \
run.googleapis.com \
dns.googleapis.com
</pre>

### Deploying

1. Navigate to `terraform` folder.
2. Customize `terraform.tfvars` based on `terraform.tfvars.template`.
3. Run: `terraform init`, `terraform plan`, `terraform apply`
5. Check the output to get the URL for your newly deployed service or get it via the following command: `gcloud run services list | grep -i diydns-function`

### Deploying changes

In case you made some customization to the code, you can easily push the changes with gcloud (from project's root folder):

<pre>
export REGION=REGION_YOU_SET_IN_TERRAFORM_TFVARS

gcloud functions deploy diydns-function \
  --gen2 \
  --region=${REGION} \
  --runtime=go122 \
  --source=. \
  --entry-point=handleHTTP \
  --trigger-http
</pre>

## Usage

1. On your home machine/server/raspberrypi (ideally the one that runs 24/7), setup a cronjob via `crontab -e`

2. Use this template: 

    `30 * * * * curl -s "https://{CLOUD_FUNCTION_URL}/?key={YOUR_SECRET_KEY}&zone={YOUR_MANAGED_DNS_ZONE}&domain={YOUR_DNS_ENTRY_NAME}." >> /tmp/diydyndns-domain.log 2>&1` 

    > :info: Yes,  there is dot at the end of the domain entry - e.g.: `subdomain.domain.tld.`

    > :info: This will run every 30 minutes, but you can adjust this number in the beginning of the line (from 30 to 60 for example).

3. Check the `/tmp/diydyndns-domain.log` to see if it's working. By default it runs every 30 minutes.

## License

Apache License 2.0. See the [LICENSE](LICENSE) file.
