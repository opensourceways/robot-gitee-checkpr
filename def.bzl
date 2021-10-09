load("@io_bazel_rules_docker//container:pull.bzl", "container_pull")
load("@io_bazel_rules_docker//container:image.bzl", "container_image")
load("@io_bazel_rules_docker//container:bundle.bzl", "container_bundle","container_bundle_")
load("@io_bazel_rules_docker//contrib:push-all.bzl", "container_push")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")

def containers():
    container_pull(
        name = "git-base",
        digest = "sha256:01b0f83fe91b782ec7ddf1e742ab7cc9a2261894fd9ab0760ebfd39af2d6ab28",  # 2018/07/02
        registry = "gcr.io",
        repository = "k8s-prow/git",
        tag = "0.2",
    )

# build_image is a macro for creating :app and :image targets
def build_image(
        name,  # use "image"
        base = None,
        stamp = True,  # stamp by default, but allow overrides
        app_name = "app",
        **kwargs):
    go_image(
        name = app_name,
        base = base,
        embed = [":go_default_library"],
        goarch = "amd64",
        goos = "linux",
    )

    container_image(
        name = name,
        base = ":" + app_name,
        stamp = stamp,
        **kwargs
    )

# push_image creates a bundle of container images, and a target to push them.
def push_image(
        name,
        bundle_name = "bundle",
        images = None):
    container_bundle(
        name = bundle_name,
        images = images,
    )
    container_push(
        name = name,
        bundle = ":" + bundle_name,
        format = "Docker",
    )

# image_tags returns a {image: target} map for each cmd or {name: target}
# Concretely,image_tags("//checkpr:image") will output the following:
# {
#   "swr.ap-southeast-1.myhuaweicloud.com/opensourceway/robot/checkpr:20210203-deadbeef": //checkpr:image
#   "swr.ap-southeast-1.myhuaweicloud.com//opensourceway/robot/checkpr:latest": //checkpr:image
# }
def image_tags(target):
    docker_name = "{STABLE_REPO}/{DOCKER_NAME}"
    outs = {}
    outs["%s:{DOCKER_TAG}" % docker_name] = target
    outs["%s:latest" % docker_name] = target
    return outs
