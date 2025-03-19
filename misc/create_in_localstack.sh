#!/bin/sh

export AWS_DEFAULT_REGION=us-west-2
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

awslocal dynamodb create-table \
  --table-name Extensions \
  --attribute-definitions AttributeName=ID,AttributeType=S \
  --key-schema AttributeName=ID,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=10,WriteCapacityUnits=10 || true

awslocal s3 mb s3://brave-core-ext || true