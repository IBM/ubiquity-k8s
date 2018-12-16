set -x
export DOCKER_REGISTRY=stg-artifactory.haifa.ibm.com:5030
export PATH=$PATH:/usr/local/go/bin:$WORKSPACE/work/bin
pwd
ls
PWDO=`pwd`
export GOPATH=`pwd`/work
repo="work/src/github.com/IBM/ubiquity-k8s"
rm -rf $repo || :
mkdir -p $repo
mv * $repo || :
cd $repo

pwd
ls


if [ -d ./deploy ]; then
   # before cleanup in the repo -> https://github.com/IBM/ubiquity-k8s/pull/201  so the acceptance in ./deploy
   tar cvf $PWDO/acceptance_tests.tar ./deploy ./scripts/*acceptance*
else
   tar cvf $PWDO/acceptance_tests.tar ./scripts/acceptance_tests
fi

echo "======================================================================="
echo "Prepare the formal TAR.GZ file...."
cd scripts
all_in_one="k8s_scbe_all_in_one"
if [ -d  "${all_in_one}" ]; then
   tar cvf $PWDO/${all_in_one}.tar ${all_in_one}
else
   all_in_one="installer-for-ibm-storage-enabler-for-containers"
   all_in_one_with_ver=${all_in_one}-${IMAGE_VERSION}-${BUILD_NUMBER}
   mv ${all_in_one} ${all_in_one_with_ver}
   find . -type f  | xargs chmod 640
   find . -type f -name "*.sh"  | xargs chmod 740
   tar czvf $PWDO/${all_in_one_with_ver}.tar.gz ${all_in_one_with_ver}
   ls -l $PWDO/${all_in_one_with_ver}.tar.gz
   file $PWDO/${all_in_one_with_ver}.tar.gz
   echo "Formal TAR.GZ file is ready -> $PWDO/${all_in_one_with_ver}.tar.gz"
   mv  ${all_in_one_with_ver} ${all_in_one}
fi
   cd -
echo "======================================================================="

echo "------- Docker images build and push - Start"

branch=`echo $GIT_BRANCH| sed 's|/|.|g'`  #not sure if docker accept / in the version
specific_tag="${IMAGE_VERSION}_b${BUILD_NUMBER}_${branch}"

if [ "$GIT_BRANCH" = "dev" -o "$GIT_BRANCH" = "origin/dev" -o "$GIT_BRANCH" = "master" -o "$to_tag_latest_also_none_dev_branches" = "true" ]; then
   tag_latest="true"
   echo "will tag latest \ version in addition to the branch tag $GIT_BRANCH"
else
   tag_latest="false"
   echo "NO latest \ version tag for you $GIT_BRANCH"
fi

echo "build ubiquity provisioner image"
echo "================================"
ubiquity_registry="${DOCKER_REGISTRY}/${UBIQUITY_K8S_PROVISIONER_IMAGE}"
ubiquity_provisioner_tag_specific="${ubiquity_registry}:${specific_tag}"
ubiquity_provisioner_tag_latest=${ubiquity_registry}:latest
ubiquity_provisioner_tag_version=${ubiquity_registry}:${IMAGE_VERSION}
[ "$tag_latest" = "true" ] && taglatestflag="-t ${ubiquity_provisioner_tag_latest} -t ${ubiquity_provisioner_tag_version}" || taglatestflag=""
# Build and tags together

docker build -t ${ubiquity_provisioner_tag_specific} $taglatestflag -f Dockerfile.Provisioner .
# push the tags
docker push ${ubiquity_provisioner_tag_specific}
[ "$tag_latest" = "true" ] && docker push ${ubiquity_provisioner_tag_latest} || :
[ "$tag_latest" = "true" ] && docker push ${ubiquity_provisioner_tag_version} || :


echo "build ubiquity flex image"
echo "========================="
ubiquity_registry="${DOCKER_REGISTRY}/${UBIQUITY_K8S_FLEX_IMAGE}"
ubiquity_flex_tag_specific="${ubiquity_registry}:${specific_tag}"
ubiquity_flex_tag_latest=${ubiquity_registry}:latest
ubiquity_flex_tag_version=${ubiquity_registry}:${IMAGE_VERSION}
[ "$tag_latest" = "true" ] && taglatestflag="-t ${ubiquity_flex_tag_latest} -t ${ubiquity_flex_tag_version}" || taglatestflag=""
# Build and tags together

docker build -t ${ubiquity_flex_tag_specific} ${taglatestflag} -f Dockerfile.Flex .
# push the tags
docker push ${ubiquity_flex_tag_specific}
[ "$tag_latest" = "true" ] && docker push ${ubiquity_flex_tag_latest} || :
[ "$tag_latest" = "true" ] && docker push ${ubiquity_flex_tag_version} || :


echo "------- Docker images build and push - Done"


cd $PWDO


echo "============================="
echo "ubiquity provisioner IMAGE name : "
echo "   specific tag : ${ubiquity_provisioner_tag_specific}"
[ "$tag_latest" = "true" ] && echo "   latest tag \ version   : ${ubiquity_provisioner_tag_latest}    ${ubiquity_provisioner_tag_version}" || echo "no latest tag"

echo "============================="
echo "ubiquity flex IMAGE name : "
echo "   specific tag : ${ubiquity_flex_tag_specific}"
[ "$tag_latest" = "true" ] && echo "   latest tag \ version   : ${ubiquity_flex_tag_latest}      ${ubiquity_flex_tag_version}"  || echo "no latest tag"

echo "============================="
echo "ubiquity hookexecutor IMAGE name : "
echo "   specific tag : ${ubiquity_hookexecutor_tag_specific}"
[ "$tag_latest" = "true" ] && echo "   latest tag \ version   : ${ubiquity_hookexecutor_tag_latest}      ${ubiquity_hookexecutor_tag_version}"  || echo "no latest tag"


echo ${ubiquity_provisioner_tag_specific} > ubiquity_k8s_tags
echo ${ubiquity_flex_tag_specific} >> ubiquity_k8s_tags
echo ${ubiquity_hookexecutor_tag_specific} >> ubiquity_k8s_tags