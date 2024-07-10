# dkr

This is a simple wrapper around the docker cli tool. When you are in a directory with a `Dockerfile` you can run `dkr build` to build the project. The image will have the same name as the directory the `Dockerfile` is in

To run the container from the image `dkr run` which will start the container, daemonised with the necessary ports published according to the `EXPOSE` commands in the file

You can connect with `dkr connect` which will open a bash shell

Finally the container can be stopped with `dkr stop` and deleted with `dkr delete`

For the confident you can build, run and connect all in one go `dkr build run connect` and each command will be executed providing the previous command did not fail. `dkr build run connect stop delete` is also a thing :P

## A little magic

`dkr [--file otherdocker.yml] build|run|connect|stop|delete`

The optional --file flag will allow you to override the Dockerfile

|Command| Description|
|---|---|
|`build`|Builds the image from Dockerfile|
|`run`|Runs the container daemonised|
|`connect`|Connect to the container|
|`stop`|Stops container|
|`delete`|Delete container|

The dockerfile can contain various extra tags that will
allow dkr to run things for you

|Tag|Description|
|---|---|
|`#RUN`|Anything after this will be passed to the run command|
|`#BUILD`|Anything after this will be passed to the build command|
|`#IGNORE`|Anything after this will be added to the .dockerignore file|
|`#NAME`|By default the container name is the same as the directory but this tag will allow you to set a name|
