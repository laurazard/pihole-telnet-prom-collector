# Copied (and modified) from https://github.com/docker/cli.
# All credits to the original authors.

target "_common" {
    args = {
        BUILDKIT_CONTEXT_KEEP_GIT_DIR = 1
    }
}

target "_platforms" {
    platforms = [
        "linux/amd64",
        "linux/arm/v6",
        "linux/arm/v7",
        "linux/arm64",
        "linux/ppc64le",
        "linux/riscv64",
        "linux/s390x",
    ]
}

group "default" {
    targets = ["binary"]
}

target "binary" {
    inherits = ["_common"]
    target = "binary"
    platforms = ["local"]
    output = ["build"]
}

target "cross" {
    inherits = ["binary", "_platforms"]
}
