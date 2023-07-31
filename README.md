# dkr

This is a simple wrapper around the docker cli tool. When you are in a directory with a `Dockerfile` you can run `dkr build` to build the project. The image will have the same name as the directory the `Dockerfile` is in

To run the container from the image `dkr run` which will start the container, daemonised with the necessary ports published according to the `EXPOSE` commands in the file

You can connect with `dkr connect` which will open a bash shell

Finally the container can be stopped with `dkr stop` and deleted with `dkr delete`

For the confident you can build, run and connect all in one go `dkr build run connect` and each command will be executed providing the previous command did not fail. `dkr build run connect stop delete` is also a thing :P

If `dokter` is installed the `lint` command will do a simple audit of your `Dockerfile`. I have included this in hopes that the audit will improve over time

If `trivy` is installed the `scan` command will provide a vulnerability report for the software in the image

## A little magic

One minor addition I have added is the `#IGNORE` command for the `Dockerfile`. This will add the arguments to the `.dockerignore` file so you can make sure that all the information needed to build a docker image is in one place

And another one is `#VOLUME xxx` which will add a `-v xxx` argument when running the container

To set a default name that is not the directory that the `Dockerfile` is in use `#NAME other_name`

To pass arguments to be used to run a container add them with `#RUN`

By default `Dockerfile` will be used if it is available in the current directory. The `--file` option will allow you to use another file

# dcctl

Another little tool to build images in a project and handle the `docker-compose.yaml` file
