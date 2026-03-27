<?php
namespace Weiyun;

class SHA1 {
    public $h0 = 0x67452301;
    public $h1 = 0xEFCDAB89;
    public $h2 = 0x98BADCFE;
    public $h3 = 0x10325476;
    public $h4 = 0xC3D2E1F0;
    private $message_byte_length = 0;
    private $unprocessed = "";

    private function _left_rotate($n, $b) {
        return (($n << $b) | ($n >> (32 - $b))) & 0xFFFFFFFF;
    }

    public function update($data) {
        $this->unprocessed .= $data;
        $this->message_byte_length += strlen($data);
        while (strlen($this->unprocessed) >= 64) {
            $this->_process_chunk(substr($this->unprocessed, 0, 64));
            $this->unprocessed = substr($this->unprocessed, 64);
        }
    }

    private function _process_chunk($chunk) {
        $w = array_fill(0, 80, 0);
        for ($i = 0; $i < 16; $i++) {
            $unpacked = unpack('N', substr($chunk, $i * 4, 4));
            $w[$i] = $unpacked[1];
        }
        for ($i = 16; $i < 80; $i++) {
            $w[$i] = $this->_left_rotate($w[$i - 3] ^ $w[$i - 8] ^ $w[$i - 14] ^ $w[$i - 16], 1);
        }
        $a = $this->h0; $b = $this->h1; $c = $this->h2; $d = $this->h3; $e = $this->h4;
        for ($i = 0; $i < 80; $i++) {
            if ($i >= 0 && $i <= 19) {
                $f = ($b & $c) | ((~$b) & $d);
                $k = 0x5A827999;
            } elseif ($i >= 20 && $i <= 39) {
                $f = $b ^ $c ^ $d;
                $k = 0x6ED9EBA1;
            } elseif ($i >= 40 && $i <= 59) {
                $f = ($b & $c) | ($b & $d) | ($c & $d);
                $k = 0x8F1BBCDC;
            } else {
                $f = $b ^ $c ^ $d;
                $k = 0xCA62C1D6;
            }
            $temp = ($this->_left_rotate($a, 5) + $f + $e + $k + $w[$i]) & 0xFFFFFFFF;
            $e = $d; $d = $c; $c = $this->_left_rotate($b, 30); $b = $a; $a = $temp;
        }
        $this->h0 = ($this->h0 + $a) & 0xFFFFFFFF;
        $this->h1 = ($this->h1 + $b) & 0xFFFFFFFF;
        $this->h2 = ($this->h2 + $c) & 0xFFFFFFFF;
        $this->h3 = ($this->h3 + $d) & 0xFFFFFFFF;
        $this->h4 = ($this->h4 + $e) & 0xFFFFFFFF;
    }

    public function get_state() {
        return bin2hex(pack('V5', $this->h0, $this->h1, $this->h2, $this->h3, $this->h4));
    }

    public function hexdigest() {
        $copy = clone $this;
        $rem = $copy->message_byte_length % 64;
        $pad = "\x80";
        if ($rem < 56) {
            $pad .= str_repeat("\x00", 56 - 1 - $rem);
        } else {
            $pad .= str_repeat("\x00", 64 - 1 - $rem + 56);
        }
        $pad .= pack('J', $copy->message_byte_length * 8);
        $copy->update($pad);
        return sprintf("%08x%08x%08x%08x%08x", $copy->h0, $copy->h1, $copy->h2, $copy->h3, $copy->h4);
    }
}

class Client {
    private $token;
    private $mcpUrl;
    private $requestId = 0;

    public function __construct($token, $mcpUrl = "https://www.weiyun.com/api/v3/mcpserver") {
        $this->token = $token;
        $this->mcpUrl = $mcpUrl;
    }

    public function call($toolName, $arguments) {
        $this->requestId++;
        $payload = [
            "jsonrpc" => "2.0",
            "id" => $this->requestId,
            "method" => "tools/call",
            "params" => [
                "name" => $toolName,
                "arguments" => $arguments
            ]
        ];

        $options = [
            'http' => [
                'header'  => "Content-type: application/json\r\nWyHeader: mcp_token={$this->token}\r\n",
                'method'  => 'POST',
                'content' => json_encode($payload),
                'timeout' => 120,
            ]
        ];
        $context  = stream_context_create($options);
        $result = @file_get_contents($this->mcpUrl, false, $context);
        if ($result === FALSE) {
            throw new \Exception("Request failed.");
        }

        $resDecoded = json_decode($result, true);
        if (isset($resDecoded['result']['content'])) {
            foreach ($resDecoded['result']['content'] as $item) {
                if (isset($item['type']) && $item['type'] === 'text') {
                    return json_decode($item['text'], true);
                }
            }
        }
        return $resDecoded;
    }

    /**
     * List files and directories.
     *
     * @param int $limit
     * @param int $getType
     * @param int $offset
     * @param string|null $dirKey
     * @param string|null $pdirKey
     */
    public function listFiles($limit = 50, $getType = 0, $offset = 0, $dirKey = null, $pdirKey = null) {
        $args = ["limit" => $limit, "get_type" => $getType, "offset" => $offset];
        if ($dirKey !== null) $args["dir_key"] = $dirKey;
        if ($pdirKey !== null) $args["pdir_key"] = $pdirKey;
        return $this->call("weiyun.list", $args);
    }

    /**
     * Get HTTPS download links for a list of items.
     *
     * @param array $items  Array of ["file_id" => ..., "pdir_key" => ...]
     */
    public function download($items) {
        return $this->call("weiyun.download", ["items" => $items]);
    }

    /**
     * Delete files and/or directories.
     *
     * @param array|null $fileList        Array of ["file_id" => ..., "pdir_key" => ...]
     * @param array|null $dirList         Array of ["dir_key" => ..., "pdir_key" => ...]
     * @param bool $deleteCompletely      If true, permanently deletes; otherwise moves to trash
     */
    public function delete($fileList = null, $dirList = null, $deleteCompletely = false) {
        $args = ["delete_completely" => $deleteCompletely];
        if ($fileList) $args["file_list"] = $fileList;
        if ($dirList) $args["dir_list"] = $dirList;
        return $this->call("weiyun.delete", $args);
    }

    /**
     * Generate a public share link.
     *
     * @param array|null $fileList   Array of ["file_id" => ..., "pdir_key" => ...]
     * @param array|null $dirList    Array of ["dir_key" => ..., "pdir_key" => ...]
     * @param string|null $shareName Custom name for the share link
     */
    public function genShareLink($fileList = null, $dirList = null, $shareName = null) {
        $args = [];
        if ($fileList) $args["file_list"] = $fileList;
        if ($dirList) $args["dir_list"] = $dirList;
        if ($shareName !== null) $args["share_name"] = $shareName;
        return $this->call("weiyun.gen_share_link", $args);
    }

    /**
     * Calculate upload parameters (chunked SHA1 states + MD5).
     *
     * @param string $filePath
     * @return array
     */
    public function calcUploadParams($filePath) {
        $fileSize = filesize($filePath);
        $blockSize = 524288;
        $lastBlockSize = $fileSize % $blockSize;
        if ($lastBlockSize == 0) $lastBlockSize = $blockSize;
        $checkBlockSize = $lastBlockSize % 128;
        if ($checkBlockSize == 0) $checkBlockSize = 128;
        $beforeBlockSize = $fileSize - $lastBlockSize;

        $sha1 = new SHA1();
        $md5Ctx = hash_init('md5');
        $blockShaList = [];

        $fp = fopen($filePath, 'rb');
        for ($offset = 0; $offset < $beforeBlockSize; $offset += $blockSize) {
            $data = fread($fp, $blockSize);
            $sha1->update($data);
            hash_update($md5Ctx, $data);
            $blockShaList[] = $sha1->get_state();
        }

        $betweenLen = $lastBlockSize - $checkBlockSize;
        $betweenData = $betweenLen > 0 ? fread($fp, $betweenLen) : "";
        if ($betweenLen > 0) {
            $sha1->update($betweenData);
            hash_update($md5Ctx, $betweenData);
        }
        $checkSha = $sha1->get_state();

        $checkDataBytes = $checkBlockSize > 0 ? fread($fp, $checkBlockSize) : "";
        if ($checkBlockSize > 0) {
            $sha1->update($checkDataBytes);
            hash_update($md5Ctx, $checkDataBytes);
        }

        $fileSha = $sha1->hexdigest();
        $fileMd5 = hash_final($md5Ctx);
        $checkData = base64_encode($checkDataBytes);

        $blockShaList[] = $fileSha;
        fclose($fp);

        return [
            "filename"       => basename($filePath),
            "file_size"      => $fileSize,
            "file_sha"       => $fileSha,
            "file_md5"       => $fileMd5,
            "block_sha_list" => $blockShaList,
            "check_sha"      => $checkSha,
            "check_data"     => $checkData,
        ];
    }

    /**
     * Upload a file using the Weiyun two-phase FTN protocol.
     *
     * @param string $filePath
     * @param string|null $pdirKey  Target directory key (null = root)
     * @param int $maxRounds        Maximum number of upload rounds
     * @return array  ["file_id" => ..., "filename" => ...]
     */
    public function upload($filePath, $pdirKey = null, $maxRounds = 50) {
        $params = $this->calcUploadParams($filePath);
        $fileSize = $params["file_size"];
        $filename = $params["filename"];

        $preUploadArgs = [
            "filename"       => $filename,
            "file_size"      => $fileSize,
            "file_sha"       => $params["file_sha"],
            "file_md5"       => $params["file_md5"],
            "block_sha_list" => $params["block_sha_list"],
            "check_sha"      => $params["check_sha"],
            "check_data"     => $params["check_data"],
        ];
        if ($pdirKey !== null) {
            $preUploadArgs["pdir_key"] = $pdirKey;
        }

        $fileData = file_get_contents($filePath);

        $roundNum = 0;
        while ($roundNum < $maxRounds) {
            $roundNum++;
            $preRsp = $this->call("weiyun.upload", $preUploadArgs);

            if (!empty($preRsp["error"])) {
                throw new \Exception("预上传失败: " . $preRsp["error"]);
            }

            if (!empty($preRsp["file_exist"])) {
                return [
                    "file_id"  => $preRsp["file_id"] ?? "",
                    "filename" => $preRsp["filename"] ?? $filename,
                ];
            }

            $chList = $preRsp["channel_list"] ?? [];
            $uk = $preRsp["upload_key"] ?? "";
            $ex = $preRsp["ex"] ?? "";

            // Find a channel with data remaining
            $ch = null;
            foreach ($chList as $c) {
                if ((int)($c["len"] ?? 0) > 0) {
                    $ch = $c;
                    break;
                }
            }

            if ($ch === null) {
                $state = (int)($preRsp["upload_state"] ?? 0);
                if ($state === 2) {
                    return [
                        "file_id"  => $preRsp["file_id"] ?? "",
                        "filename" => $preRsp["filename"] ?? $filename,
                    ];
                }
                throw new \Exception("无可上传通道，upload_state={$state}");
            }

            $offset    = (int)$ch["offset"];
            $length    = (int)$ch["len"];
            $channelId = (int)$ch["id"];
            $actualLen = min($length, strlen($fileData) - $offset);

            $chunk    = substr($fileData, $offset, $actualLen);
            $chunkB64 = base64_encode($chunk);

            $cl = [];
            foreach ($chList as $c) {
                $cl[] = ["id" => (int)$c["id"], "offset" => (int)$c["offset"], "len" => (int)$c["len"]];
            }

            $upRsp = $this->call("weiyun.upload", [
                "filename"     => $filename,
                "file_size"    => $fileSize,
                "file_sha"     => $params["file_sha"],
                "block_sha_list" => [],
                "check_sha"    => $params["check_sha"],
                "upload_key"   => $uk,
                "channel_list" => $cl,
                "channel_id"   => $channelId,
                "ex"           => $ex,
                "file_data"    => $chunkB64,
            ]);

            if (!empty($upRsp["error"])) {
                throw new \Exception("分片上传失败: " . $upRsp["error"]);
            }

            $state = (int)($upRsp["upload_state"] ?? 0);
            if ($state === 2) {
                return [
                    "file_id"  => $upRsp["file_id"] ?? "",
                    "filename" => $upRsp["filename"] ?? $filename,
                ];
            }
        }

        throw new \Exception("超过最大上传轮数 ({$maxRounds})，上传未完成");
    }
}
