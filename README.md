# Adabas administration client

This code contains a sample use case of Adabas RESTful administration.
The Swagger definition delivered in the RESTful server can be used to
generate a RESTful client.

The example shows, how Adabas administration tasks can be generatored using swagger generators.
This examples uses the programming language GO but you can use other programming languages like Angular, Python or specific Java frameworks as well.

The result is a command line access to the Adabas RESTful administration.

## Build

The build process needs an installed GO (golang) compiler. The build works on Linux. The build-infrastructure of GO requires web access to download required dependencies.

The command line is generated using

```bash
make
```

You can regenerate sources out of swagger definitions by using

```bash
make generate
```

Except the `cmd` directory, all other directories
contain generated source.

## Generate binaries

The final binary is generated using the

```bash
make
```

command. The binary are located in the corresponding `bin/${GOOS}_${GOARCH}` directory. You can create cross operating system binaries. On Linux the `make` command will generate a Linux binary call `client`.

If you'd like to generate a Windows binary, you need to provide an additional environment variable `GOOS`. For example to build for Windows, following command need to be entered.

```bash
make GOOS=windows
```

## Runtime

Beside the direct usage of the client you might use the `startAdmin.sh` script for a quick start.  The command line tool provided all help entering the  `help` command.

The client has a `-url` option.
This option can be used to reference the REST server location. It is possible to use `<host>:<port>` to specify a HTTP access. To connect to an HTTPS connection, the URL need to specifiy the SSL connection with `https://<host>:<url>`. A preset URL can be set using the environment variable `ADABAS_ADMIN_URL`. To omit entering the password for each request, you can set the environment `ADABAS_ADMIN_PASSWORD`.

If the certificate is for internal use without public certification, you may switch off validation using the `-ignoreTLS` switch.

With `-dbid` and `-fnr` you may define corresponding database parameter. Dynamic parameters and input definitions are entered using the `-param` or `-input` options.

### Example of Parameter usage

You may provide parameter to perform special operations. For example to set new Adabas parameters, the parameter need to be passed using the `-param` parameter. Example to set new parameters in the static Adabas parameter definition

```sh
client -url localhost:8120 -dbid 24 -param type=static,PLOG=YES,NT=5 setparameter
```

This example will set new Adabas static paramters for the database 24.

### Create Adabas database

To create a new Adabas database, use a input file with the JSON definition of the new database. Environment variables will be resolved on the remote RESTful server.

```JSON
{
    "CheckpointFile":1,
    "ContainerList":[        {"BlockSize":"8K","ContainerSize":"20M","Path":"${ADADATADIR/db075/ASSO1.075"}, {"BlockSize":"32K","ContainerSize":"20M","Path":"${ADADATADIR}/db075/ASSO2.075"},        {"BlockSize":"32K","ContainerSize":"20M","Path":"${ADADATADIR}/db075/DATA1.075"},{"BlockSize":"16K","ContainerSize":"20M","Path":"${ADADATADIR}/db075/WORK.075"}
    ],
    "Dbid":75,
    "LoadDemo":true,
    "Name":"DEMODB",
    "SecurityFile":2,
    "UserFile":3
}
```

The corresponding JSON file need to be referenced using the `-input` option.

### Create Adabas file

To create a new Adabas file, use a input file with the JSON definition of the new Adabas file

```JSON
{
    "fileNumber":350,
    "fduOptions":{
        "fduName":"GO_TEST",
    },
    "fdtDefinition":"1,AQ%2,AF,15,A,NU%1,NN,20,A,DE,UQ%1,VN,20,A,DE"
}

```

The corresponding JSON file need to be referenced using the `-input` option.

