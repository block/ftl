# EKS Example

This directory contains some example code to bootstrap an EKS cluster suitable for use by FTL.

Note that EKS and FTL should be initialized separately, as the kubernetes based helm TF provider should not be
part of the same state as the EKS cluster.

