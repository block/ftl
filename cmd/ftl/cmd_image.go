package main

type imageCmd struct {
	Build   imageBuildCmd   `cmd:"" help:"Build an image from the modules in the project."`
	Inspect imageInspectCmd `cmd:"" help:"Inspect the schema of an image."`
	Deploy  imageDeployCmd  `cmd:"" help:"Deploy images to the cluster."`
}
