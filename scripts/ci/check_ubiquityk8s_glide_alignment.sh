# this script will fail the ci build if glide.yml of ubiquity-k8s is not aligned with ubiquity latest commit

set +x

echo "Fail k82 images build if git is not aligned with ubiquity"
latest_commit_on_ubiquity_dev=`git ls-remote https://github.com/IBM/ubiquity | grep dev$ | awk '{ print $1 }'`

latest_commit_on_glide_yaml=`cat glide.yaml | grep IBM/ubiquity -a1 | grep version | awk '{ print $2 }'`

if [ $latest_commit_on_ubiquity_dev != $latest_commit_on_glide_yaml ]; then
    echo "Latest commit on github.com/IBM/ubiquity dev: $latest_commit_on_ubiquity_dev"
    echo "Latest commit on glide.yaml: $latest_commit_on_glide_yaml"
	exit 1
fi
exit 0
