set +x

echo "Fail k82 images build if git is not aligned with ubiquity"
latest_commit_on_ubiquity_dev=`git ls-remote https://github.com/IBM/ubiquity | grep dev$ | awk '{ print $1 }'`
echo "Latest commit on github.com/IBM/ubiquity dev: $latest_commit_on_ubiquity_dev"

latest_commit_on_glide_yaml=`cat glide.yaml | grep IBM/ubiquity -a1 | grep version | awk '{ print $2 }'`
echo "Latest commit on glide.yaml: $latest_commit_on_glide_yaml"

if [ $latest_commit_on_ubiquity_dev != $latest_commit_on_glide_yaml ]; then
	exit 1
fi
exit 0
