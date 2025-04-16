load("@bazel_gazelle//:def.bzl", "gazelle")
load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

# gazelle:prefix github.com/block/ftl
gazelle(name = "gazelle")

go_library(
    name = "ftl",
    srcs = ["version.go"],
    importpath = "github.com/block/ftl",
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_alecthomas_types//must",
        "@org_golang_x_mod//semver",
    ],
)

go_test(
    name = "ftl_test",
    srcs = ["version_test.go"],
    embed = [":ftl"],
    deps = ["@com_github_alecthomas_assert_v2//:assert"],
)
