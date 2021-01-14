# Waves

![Go](https://github.com/wujie1993/waves/workflows/Go/badge.svg?branch=main)

Waves is an application deployment platform, mainly used for offline fast delivery of applications. Its design is inspired by [kubernetes](https://github.com/kubernetes/kubernetes), uses a declarative interface to manage resources and provides simple command line management tools.

## Dependency

- golang 1.13
- etcd 3.3.19
- ansible
- mitogen

**Install Dependency**

```bash
// install ansible and mitogen
yum install -y ansible python-pip
pip install mitogen

// install swag
git clone -b v1.6.7 https://github.com/swaggo/swag.git $GOPATH/src/github.com/swaggo/swag
go install github.com/swaggo/swag/cmd/swag
```

## Build

```
make
```

## Run

1. Run ETCD for storing data
   
   ```bash
   docker run -d -p 2379:2379 --rm quay.io/coreos/etcd:v3.3.19 /usr/local/bin/etcd --listen-client-urls http://0.0.0.0:2379 --advertise-client-urls http://0.0.0.0:2379
   ```

2. Edit config file

   ```bash
   cp conf/app.ini.sample conf/app.ini
   ```

3. Run as frontend service
   
   ```bash
   make run
   ```

3. Access via browser

   ```
   http://localhost:8000/deployer/swagger/index.html
   ```

## Others

**generate swagger api doc**

```
make doc
```

**generate codes**

```
make gen
```

**unit test**

```
make test
```

**clean working directory**

```
make clean
```

## Docs

- [设计文档 | 调度器](./pkg/schedule/README.md)
- [设计文档 | 管理器](./pkg/operators/README.md)
- [开发文档](./docs/Develop.md)
- [参与项目](./CONTRIBUTING.md)
