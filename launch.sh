#!/bin/sh

#
# Changes these settings
#

# number of instances to start up
COUNT=3

# type of instances to start
TYPE=c3.large

# See the CoreOS AMIs here:
# https://coreos.com/docs/running-coreos/cloud-providers/ec2/
#
# You probably want to use the 'HVM' images
#
REGION=us-west-2
AMI=ami-f52c63c5

# Your ec2 key pair name
KEY_NAME=james-mac

# your security group - should probably
# allow open TCP access to other hosts in the same security group
SECURITY_GROUP=sg-79d6b71c

# subnet ID for your VPC
SUBNET_ID=subnet-fba14190


# Probably don't need to change this:
#
# uses the python aws cli
#
USER_DATA=`cat launch-user-data.yml`

aws ec2 run-instances \
  --region $REGION    \
  --image-id $AMI     \
  --count $COUNT      \
  --instance-type $TYPE      \
  --key-name $KEY_NAME          \
  --security-group-id $SECURITY_GROUP  \
  --subnet-id $SUBNET_ID \
  --user-data "$USER_DATA"
