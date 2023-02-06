# Treehole_shamir_client

解密用户邮箱的客户端，opentreehole密钥管理员使用。

## 使用步骤

必要：准备 pgp 私钥。

如果使用 gnupg 生成密钥，请从 gnupg 导出私钥文本文件，并放置到程序同一目录下。

如果使用其他工具生成密钥，请自行处理

```shell
gpg -a -o private.key --export-private-keys <your_uid>
```

### 解密 shamir 信息

```shell
# 解密全部用户信息，并且上传到后端等待解密
shamir_client decrypt -k <your_private_key_file> -a <server_url> -p <your_password>

# 解密单用户的信息
shamir_client decrypt -k <your_private_key_file> -a <server_url>  -p <your_password> -u <user_id>
```

### 生成公私钥文件

```shell
shamir_client generate <your_name> <your_email> <your_password>
```

### 解密用户邮箱

准备 `shares.json` 文件，放置到程序同一目录下，格式为

```json
[
  "123123123\n456456456",
  "345345345\n678678678"
]
```

执行

```shell
shamir_client email [-f <your_share_file>]
```