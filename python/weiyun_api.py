import base64
import hashlib
import json
import os
import struct
import requests

def _left_rotate(n, b):
    return ((n << b) | (n >> (32 - b))) & 0xFFFFFFFF

class SHA1:
    def __init__(self):
        self.h0 = 0x67452301
        self.h1 = 0xEFCDAB89
        self.h2 = 0x98BADCFE
        self.h3 = 0x10325476
        self.h4 = 0xC3D2E1F0
        self._message_byte_length = 0
        self._unprocessed = b""

    def update(self, data):
        self._unprocessed += data
        self._message_byte_length += len(data)
        while len(self._unprocessed) >= 64:
            self._process_chunk(self._unprocessed[:64])
            self._unprocessed = self._unprocessed[64:]

    def _process_chunk(self, chunk):
        assert len(chunk) == 64
        w = [0] * 80
        for i in range(16):
            w[i] = struct.unpack(">I", chunk[i * 4:(i + 1) * 4])[0]
        for i in range(16, 80):
            w[i] = _left_rotate(w[i - 3] ^ w[i - 8] ^ w[i - 14] ^ w[i - 16], 1)
        a, b, c, d, e = self.h0, self.h1, self.h2, self.h3, self.h4
        for i in range(80):
            if 0 <= i <= 19:
                f = (b & c) | ((~b) & d)
                k = 0x5A827999
            elif 20 <= i <= 39:
                f = b ^ c ^ d
                k = 0x6ED9EBA1
            elif 40 <= i <= 59:
                f = (b & c) | (b & d) | (c & d)
                k = 0x8F1BBCDC
            elif 60 <= i <= 79:
                f = b ^ c ^ d
                k = 0xCA62C1D6
            temp = (_left_rotate(a, 5) + f + e + k + w[i]) & 0xFFFFFFFF
            e = d
            d = c
            c = _left_rotate(b, 30)
            b = a
            a = temp
        self.h0 = (self.h0 + a) & 0xFFFFFFFF
        self.h1 = (self.h1 + b) & 0xFFFFFFFF
        self.h2 = (self.h2 + c) & 0xFFFFFFFF
        self.h3 = (self.h3 + d) & 0xFFFFFFFF
        self.h4 = (self.h4 + e) & 0xFFFFFFFF

    def get_state(self):
        assert len(self._unprocessed) == 0
        result = b""
        for h in (self.h0, self.h1, self.h2, self.h3, self.h4):
            result += struct.pack("<I", h)
        return result.hex()

    def hexdigest(self):
        message_byte_length = self._message_byte_length
        unprocessed = self._unprocessed
        h0, h1, h2, h3, h4 = self.h0, self.h1, self.h2, self.h3, self.h4
        unprocessed += b"\x80"
        unprocessed += b"\x00" * ((56 - len(unprocessed) % 64) % 64)
        unprocessed += struct.pack(">Q", message_byte_length * 8)
        tmp = SHA1.__new__(SHA1)
        tmp.h0, tmp.h1, tmp.h2, tmp.h3, tmp.h4 = h0, h1, h2, h3, h4
        tmp._unprocessed = b""
        tmp._message_byte_length = message_byte_length
        while len(unprocessed) >= 64:
            tmp._process_chunk(unprocessed[:64])
            unprocessed = unprocessed[64:]
        return "{:08x}{:08x}{:08x}{:08x}{:08x}".format(
            tmp.h0, tmp.h1, tmp.h2, tmp.h3, tmp.h4)

BLOCK_SIZE = 524288

def calc_upload_params(file_path):
    file_size = os.path.getsize(file_path)
    filename = os.path.basename(file_path)
    last_block_size = file_size % BLOCK_SIZE
    if last_block_size == 0:
        last_block_size = BLOCK_SIZE
    check_block_size = last_block_size % 128
    if check_block_size == 0:
        check_block_size = 128
    before_block_size = file_size - last_block_size
    block_sha_list = []
    sha1 = SHA1()
    md5 = hashlib.md5()
    with open(file_path, "rb") as f:
        for offset in range(0, before_block_size, BLOCK_SIZE):
            data = f.read(BLOCK_SIZE)
            sha1.update(data)
            md5.update(data)
            block_sha_list.append(sha1.get_state())
        between_data = f.read(last_block_size - check_block_size)
        sha1.update(between_data)
        md5.update(between_data)
        check_sha = sha1.get_state()
        check_data_bytes = f.read(check_block_size)
        sha1.update(check_data_bytes)
        md5.update(check_data_bytes)
        file_sha = sha1.hexdigest()
        check_data = base64.b64encode(check_data_bytes).decode("utf-8")
        block_sha_list.append(file_sha)
    file_md5 = md5.hexdigest()
    return {
        "filename": filename,
        "file_size": file_size,
        "file_sha": file_sha,
        "file_md5": file_md5,
        "block_sha_list": block_sha_list,
        "check_sha": check_sha,
        "check_data": check_data,
    }


class WeiyunClient:
    def __init__(self, token: str, mcp_url: str = "https://www.weiyun.com/api/v3/mcpserver"):
        self.token = token
        self.mcp_url = mcp_url
        self._request_id = 0

    def _call(self, tool_name: str, arguments: dict):
        self._request_id += 1
        headers = {
            "Content-Type": "application/json",
            "WyHeader": f"mcp_token={self.token}",
        }
        payload = {
            "jsonrpc": "2.0",
            "id": self._request_id,
            "method": "tools/call",
            "params": {"name": tool_name, "arguments": arguments},
        }
        resp = requests.post(self.mcp_url, headers=headers, json=payload, timeout=120)
        resp.raise_for_status()
        result = resp.json()
        content = result.get("result", {}).get("content", [])
        for item in content:
            if item.get("type") == "text":
                return json.loads(item["text"])
        return result

    def list_files(self, limit: int = 50, get_type: int = 0, offset: int = 0, dir_key: str = None, pdir_key: str = None):
        args = {"limit": limit, "get_type": get_type, "offset": offset}
        if dir_key: args["dir_key"] = dir_key
        if pdir_key: args["pdir_key"] = pdir_key
        return self._call("weiyun.list", args)

    def download(self, items: list):
        return self._call("weiyun.download", {"items": items})

    def delete(self, file_list: list = None, dir_list: list = None, delete_completely: bool = False):
        args = {"delete_completely": delete_completely}
        if file_list: args["file_list"] = file_list
        if dir_list: args["dir_list"] = dir_list
        return self._call("weiyun.delete", args)

    def gen_share_link(self, file_list: list = None, dir_list: list = None, share_name: str = None):
        args = {}
        if file_list: args["file_list"] = file_list
        if dir_list: args["dir_list"] = dir_list
        if share_name: args["share_name"] = share_name
        return self._call("weiyun.gen_share_link", args)

    def upload(self, file_path: str, pdir_key: str = None, max_rounds: int = 50):
        params = calc_upload_params(file_path)
        file_size = params["file_size"]
        filename = params["filename"]
        
        pre_upload_args = {
            "filename": filename,
            "file_size": file_size,
            "file_sha": params["file_sha"],
            "file_md5": params["file_md5"],
            "block_sha_list": params["block_sha_list"],
            "check_sha": params["check_sha"],
            "check_data": params["check_data"],
        }
        if pdir_key:
            pre_upload_args["pdir_key"] = pdir_key

        with open(file_path, "rb") as f:
            file_data = f.read()

        round_num = 0
        while round_num < max_rounds:
            round_num += 1
            pre_rsp = self._call("weiyun.upload", pre_upload_args)
            if pre_rsp.get("error"):
                raise RuntimeError(f"预上传失败: {pre_rsp['error']}")
            
            if pre_rsp.get("file_exist", False):
                return {"file_id": pre_rsp.get("file_id", ""), "filename": pre_rsp.get("filename", filename)}
            
            ch_list = pre_rsp.get("channel_list", [])
            uk = pre_rsp.get("upload_key", "")
            ex = pre_rsp.get("ex", "")
            
            ch = None
            for c in ch_list:
                if int(c.get("len", 0)) > 0:
                    ch = c
                    break
            
            if ch is None:
                state = int(pre_rsp.get("upload_state", 0))
                if state == 2:
                    return {"file_id": pre_rsp.get("file_id", ""), "filename": pre_rsp.get("filename", filename)}
                raise RuntimeError(f"无可上传通道，upload_state={state}")
            
            offset = int(ch["offset"])
            length = int(ch["len"])
            channel_id = int(ch["id"])
            actual_len = min(length, len(file_data) - offset)
            
            chunk = file_data[offset:offset + actual_len]
            chunk_b64 = base64.b64encode(chunk).decode("utf-8")
            cl = [{"id": int(c["id"]), "offset": int(c["offset"]), "len": int(c["len"])} for c in ch_list]
            
            up_rsp = self._call("weiyun.upload", {
                "filename": filename,
                "file_size": file_size,
                "file_sha": params["file_sha"],
                "block_sha_list": [],
                "check_sha": params["check_sha"],
                "upload_key": uk,
                "channel_list": cl,
                "channel_id": channel_id,
                "ex": ex,
                "file_data": chunk_b64,
            })
            
            if up_rsp.get("error"):
                raise RuntimeError(f"分片上传失败: {up_rsp['error']}")
            
            state = int(up_rsp.get("upload_state", 0))
            if state == 2:
                return {"file_id": up_rsp.get("file_id", ""), "filename": up_rsp.get("filename", filename)}
                
        raise RuntimeError(f"超过最大上传轮数 ({max_rounds})，上传未完成")
