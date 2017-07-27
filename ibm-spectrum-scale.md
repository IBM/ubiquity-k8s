# IBM Spectrum Scale

* [Deployment Prerequisities](#deployment-prerequisities)
* [Configuration](#configuring-ubiquity-docker-plugin-with-ubiquity-and-spectrum-scale)
* [Volume Creation](#volume-creation-using-spectrum-scale-storage-system)
  * [Fileset Volumes](#creating-fileset-volumes)
  * [Independent Fileset Volumes](#creating-independent-fileset-volumes)
  * [Lightweight Volumes](#creating-lightweight-volumes)
  * [Fileset with Quota Volumes](#creating-fileset-with-quota-volumes)

## Deployment Prerequisities
 * Spectrum-Scale - Ensure the Spectrum Scale client (NSD client) is installed and part of a Spectrum Scale cluster.
 * NFS - Ensure hosts support mounting NFS file systems.
 
## Configuring Ubiquity Docker Plugin with Ubiquity and Spectrum Scale
 
 The following snippet shows a sample configuration file(ubiquity-client.conf):
 
 ```toml
 logPath = "/tmp/ubiquity"            # The Ubiquity Docker Plugin will write logs to file "ubiquity-docker-plugin.log" in this path.
 backends = ["spectrum-scale"]        # The Storage system backend to be used with Ubiquity to create and manage volumes. In this we configure Docker plugin to create volumes using IBM Spectrum Scale storage system.
 
 [DockerPlugin]
 port = 9000                                # Port to serve docker plugin functions
 pluginsDirectory = "/etc/docker/plugins/"  # Point to the location of the configured Docker plugin directory (create if not already created by Docker)
 
 
 [UbiquityServer]
 address = "UbiquityServiceHostname"  # IP/hostname of the Ubiquity Service
 port = 9999            # TCP port on which the Ubiquity Service is listening
 
 [SpectrumNfsRemoteConfig]  # Only relevant for use with "spectrum-scale-nfs" backend.
 ClientConfig = "192.0.2.0/20(Access_Type=RW,Protocols=3:4);198.51.100.0/20(Access_Type=RO,Protocols=3:4,Transports=TCP:UDP)"    # Mandatory. Declares the client specific settings for NFS volume exports. Access will be limited to the specified client subnet(s) and protocols.
 ```
 
## Volume Creation using Spectrum Scale Storage system

### Creating Fileset Volumes

Create a fileset volume named demo1,  using volume driver, on the gold Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo1 --opt filesystem=gold --opt backend=spectrum-scale
```

Alternatively, we can create the same volume demo1 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo1 --opt type=fileset --opt filesystem=gold --opt backend=spectrum-scale
```

Similarly, to create a fileset volume named demo2, using nfs volume driver, on the silver Spectrum Scale file system :

```bash
docker volume create -d ubiquity --opt backend=spectrum-scale-nfs --name demo2 --opt filesystem=silver
```

Create a fileset volume named demo3, using volume driver, on the default existing Spectrum Scale filesystem :

```bash
docker volume create -d ubiquity --name demo3 --opt backend=spectrum-scale
```

Create a fileset volume named demo4, using volume driver and an existing fileset modelingData, on the gold Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo4 --opt fileset=modelingData --opt filesystem=gold --opt backend=spectrum-scale
```

Alternatively, we can create the same volume named demo4 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo4 --opt type=fileset --opt fileset=modelingData --opt filesystem=gold --opt backend=spectrum-scale
```

### Creating Independent Fileset Volumes

Create an independent fileset volume named demo5, using volume driver, on the gold Spectrum Scale file system

```bash
docker volume create -d ubiquity --name demo5 --opt type=fileset --opt filesystem=gold --opt fileset-type=independent
```

Create an independent fileset volume named demo6 having an inode limit of 1024, using volume driver, on the gold Spectrum Scale file system

```bash
docker volume create -d spectrum-scale --name demo6 --opt type=fileset --opt filesystem=gold --opt fileset-type=independent --opt inode-limit=1024
```

### Creating Lightweight Volumes

Create a lightweight volume named demo7, using volume driver, within an existing fileset 'LtWtVolFileset' in the gold Spectrum Scale filesystem :

```bash
docker volume create -d ubiquity --name demo7 --opt type=lightweight --opt fileset=LtWtVolFileset --opt filesystem=gold --opt backend=spectrum-scale
```

Create a lightweight volume named demo8, using volume driver, within an existing fileset 'LtWtVolFileset' having a sub-directory 'dir1' in the gold Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo8 --opt fileset=LtWtVolFileset --opt directory=dir1 --opt filesystem=gold --opt backend=spectrum-scale
```

Alternatively, we can create the same volume named demo8 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo8 --opt type=lightweight --opt fileset=LtWtVolFileset --opt directory=dir1 --opt filesystem=gold --opt backend=spectrum-scale
```

### Creating Fileset With Quota Volumes

Create a fileset with quota volume named demo9, using volume driver, with a quota limit of 1GB in the silver Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo9 --opt quota=1G --opt filesystem=silver --opt backend=spectrum-scale
```

Alternatively, we can create the same volume named demo9 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo9 --opt type=fileset --opt quota=1G --opt filesystem=silver --opt backend=spectrum-scale
```

Create a fileset with quota volume named demo10, using volume driver and an existing fileset 'filesetQuota' having a quota limit of 1G, in the silver Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo10 --opt fileset=filesetQuota --opt quota=1G --opt filesystem=silver --opt backend=spectrum-scale
```

Alternatively, we can also create the same volume named demo10 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo10 --opt type=fileset --opt fileset=filesetQuota --opt quota=1G --opt filesystem=silver --opt backend=spectrum-scale
```
