# Load Balancer for deeplx Translation Services
# 用于 deeplx 的负载均衡翻译服务


This program is a load balancer designed to handle requests across multiple deeplx translation service endpoints, configured via a JSON file.

这个程序是一个设计用于处理跨多个 deeplx 翻译服务端点的请求的负载均衡器，通过 JSON 文件进行配置。

## Features | 功能特点
- Configuration through a JSON file
- Dynamic endpoint handling
- Automatic retry mechanism
- 通过 JSON 文件进行配置
- 动态端点处理
- 自动重试机制

## Configuration | 配置
Create a config.json file with your endpoints and tokens:
创建一个包含您的端点和令牌的 config.json 文件：

```json
{
    "token": "global_token (if any)",
    "endpoints": [
        {
            "url": "http://example1.com",
            "token": "token1"
        },
        {
            "url": "http://example2.com",
            "token": "token2"
        }
    ]
}
```

## Usage | 使用方法

### Direct run the program | 直接运行程序

Run the program with the -config flag followed by your configuration file path.

使用 -config 标志运行程序，后面跟上您的配置文件路径。

```bash
go run main.go -config ./config.json
```

### Using Docker | 基于 Docker 进行使用

An official docker image was provided: `nerdneils/deeplx-load-balancer`.
提供了一个官方的 docker 镜像: `nerdneils/deeplx-load-balancer`。

```bash
docker run -it -v ${PWD}/config.json:/etc/deeplx-load-balancer-config.json -p 1188:1188 nerdneils/deeplx-load-balancer
```

### Using docker-compose | 基于 docker-compose 进行使用

An `docker-compose.yml` file for using Traefik was provided, you could download it from [docker-compose.yml](https://github.com/nerdneilsfield/deeplx-load-balancer/blob/master/docker-compose.yml).

提供了一个基于 `traefik` 的 `docker-compose.yml` 文件，可以在这里下载 [docker-compose.yml](https://github.com/nerdneilsfield/deeplx-load-balancer/blob/master/docker-compose.yml)。


## Contributing | 贡献
Contributions, issues, and feature requests are welcome. Feel free to check issues page.

欢迎提供贡献、问题和功能请求。请随时查看问题页面。

## License

```
MIT License
Copyright (c) 2023 DengQi

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```