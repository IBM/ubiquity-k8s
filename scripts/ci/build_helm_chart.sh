#!/bin/bash -e

set -x
set -e

# update_chart_version will add the build number to the chart version
# for example: 1.0.0 -> 1.0.0-648
function update_chart_version()
{
  if [ !$PRODUCTION_BUILD ]; then
    chart_file="$CHART_PATH"Chart.yaml
    sed -i -r "s/^version: [0-9]+\.[0-9]+\.[0-9]+$/&-$BUILD_NUMBER/" $chart_file
  fi
}

function cleanup_helm()
{
	rm -rf ~/.helm
}

PRODUCTION_BUILD=false

# ci_user=
# ci_password=

CHART_REPOSITORY="https://stg-artifactory.haifa.ibm.com/artifactory/chart-repo"
CHART_REPOSITORY_NAME="artifactory"
CHART_INDEX="artifactory-charts"
CHART_FOLDER=$CHART_INDEX
INDEX_PATH="$CHART_REPOSITORY/index.yaml"
CHART_NAME="ibm-storage-enabler-for-containers"

PROJECT_ROOT=`pwd`
HELM_PATH="$PROJECT_ROOT/scripts/ci"
export PATH=$PATH:$HELM_PATH
CHART_PATH="$PROJECT_ROOT/helm_chart/$CHART_NAME/"

# load artifactory info, like ci_user and ci_password
. site_vars

update_chart_version

# init helm
helm init --client-only

# add ubiquity helm repo
helm repo add $CHART_REPOSITORY_NAME $CHART_REPOSITORY

mkdir $CHART_FOLDER

# add ubiquity helm index
helm repo index $CHART_INDEX --url $CHART_REPOSITORY

# download index.yaml
wget $INDEX_PATH
mv index.yaml $CHART_FOLDER

# package ubiquity helm chart
helm package $CHART_PATH
CHART_NAME_TGZ=`ls $CHART_NAME*`
mv $CHART_NAME_TGZ $CHART_FOLDER

# merge index.yaml
helm repo index --merge "$CHART_FOLDER/index.yaml" --url $CHART_REPOSITORY $CHART_FOLDER

# upload chart and new index
curl -u $ci_user:$ci_password -T "$CHART_FOLDER/index.yaml" "$CHART_REPOSITORY/"
curl -u $ci_user:$ci_password -T "$CHART_FOLDER/$CHART_NAME_TGZ" "$CHART_REPOSITORY/"

cleanup_helm

CHART_VERSION=`echo "$CHART_NAME_TGZ" | egrep "[0-9]+\.[0-9]+\.[0-9]+-[0-9]+" -o`
echo "Repository: $CHART_REPOSITORY" > ubiquity_helm_info
echo "Name: $CHART_NAME" >> ubiquity_helm_info
echo "Version: $CHART_VERSION" >> ubiquity_helm_info