
APPNAME := rtfblog${SUFFIX}
NAME    := rtfb/${APPNAME}
TAG     := $$(git log -1 --pretty=%h)
IMG     := ${NAME}:${TAG}
LATEST  := ${NAME}:latest

# builds the docker image
.PHONY: dbuild
dbuild:
	docker build -t ${IMG} .
	docker tag ${IMG} ${LATEST}

# save the image to a file
.PHONY: dsave
dsave:
	docker save -o ${APPNAME}.tar ${LATEST}

# runs the container
.PHONY: drun
drun:
	docker run -it --name ${APPNAME} --rm \
    --mount type=bind,source="$(shell pwd)",target=/host \
    --net=host ${LATEST}

# override entrypoint to gain interactive shell
.PHONY: dshell
dshell:
	docker run --entrypoint /bin/bash -it --name ${APPNAME} --rm \
    --mount type=bind,source="$(shell pwd)",target=/host \
    --net=host ${LATEST}

.PHONY: dclean
dclean:
	docker image rm ${LATEST}
