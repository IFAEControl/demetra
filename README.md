# demetra

This tools aims to simplify the process of building an image for the different projects for which we use yocto. The idea is to have a toml config file for project with different metadata like machine type, meta-layers, packages that have to be installed and so on. 

## Differences with petalinux

Petalinux provides an obscure (closed source) layer which tries to simplify the process of building an image for a board. This is good for first time users which don't need much flexibility but it puts more strain when the same setup have to be used by different people or in different machines. 
This project allows to use a known environment across all the development team and automatize some tasks to speed up both development and testing. 

## Support

Version 0.0.3 supports up to gatesgarth yocto branch
