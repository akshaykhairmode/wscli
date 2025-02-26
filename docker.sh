echo "$DOCKER_PASSWORD" | docker login -u akshaykhairmode --password-stdin
tag=$(git describe --tags --abbrev=0)
docker build --build-arg GIT_TAG=$tag -t akshaykhairmode/wscli:$tag -t akshaykhairmode/wscli:latest .
docker push akshaykhairmode/wscli:$tag