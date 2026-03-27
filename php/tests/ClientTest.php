<?php
require_once __DIR__ . '/../src/Weiyun/Client.php';

use PHPUnit\Framework\TestCase;
use Weiyun\Client;
use Weiyun\SHA1;

/**
 * A testable subclass that intercepts call() to return mock data.
 */
class MockClient extends Client {
    private $mockResponses = [];

    public function __construct($responses = []) {
        parent::__construct("TEST_TOKEN");
        $this->mockResponses = $responses;
    }

    public function call($toolName, $arguments) {
        return array_shift($this->mockResponses) ?? [];
    }
}

class ClientTest extends TestCase {

    public function testInstantiation() {
        $client = new Client("TEST_TOKEN");
        $this->assertInstanceOf(Client::class, $client);
    }

    public function testSha1State() {
        $sha1 = new SHA1();
        $this->assertEquals("0123456789abcdeffedcba9876543210f0e1d2c3", $sha1->get_state());

        $sha1->update("test content");
        $this->assertNotNull($sha1->hexdigest());
    }

    public function testListFiles() {
        $expected = ["file_list" => [["file_id" => "123"]], "dir_list" => []];
        $client = new MockClient([$expected]);
        $res = $client->listFiles(10);
        $this->assertArrayHasKey("file_list", $res);
        $this->assertEquals("123", $res["file_list"][0]["file_id"]);
    }

    public function testDownload() {
        $expected = ["items" => [["file_id" => "123", "https_download_url" => "https://example.com/file"]]];
        $client = new MockClient([$expected]);
        $res = $client->download([["file_id" => "123", "pdir_key" => "dir_abc"]]);
        $this->assertArrayHasKey("items", $res);
        $this->assertGreaterThan(0, count($res["items"]));
        $this->assertArrayHasKey("https_download_url", $res["items"][0]);
    }

    public function testDelete() {
        $expected = ["freed_index_cnt" => 1];
        $client = new MockClient([$expected]);
        $res = $client->delete(
            [["file_id" => "123", "pdir_key" => "dir_abc"]],
            null,
            true
        );
        $this->assertArrayHasKey("freed_index_cnt", $res);
        $this->assertGreaterThanOrEqual(1, $res["freed_index_cnt"]);
    }

    public function testGenShareLink() {
        $expected = ["short_url" => "https://share.weiyun.com/abc123"];
        $client = new MockClient([$expected]);
        $res = $client->genShareLink(
            [["file_id" => "123", "pdir_key" => "dir_abc"]],
            null,
            "My Share"
        );
        $this->assertArrayHasKey("short_url", $res);
    }

    public function testUploadFileExist() {
        // Mock: server reports file already exists (dedup)
        $expected = ["file_exist" => true, "file_id" => "existing_123", "filename" => "test.txt"];
        $client = new MockClient([$expected]);

        // Create a small temp file
        $tmpFile = tempnam(sys_get_temp_dir(), "weiyun_test_");
        file_put_contents($tmpFile, "Hello Weiyun!");

        try {
            $res = $client->upload($tmpFile);
            $this->assertArrayHasKey("file_id", $res);
            $this->assertEquals("existing_123", $res["file_id"]);
        } finally {
            unlink($tmpFile);
        }
    }

    public function testUploadChunkedCompletes() {
        // Mock: first call returns a channel, second call returns upload_state=2 (done)
        $tmpFile = tempnam(sys_get_temp_dir(), "weiyun_test_");
        file_put_contents($tmpFile, str_repeat("A", 256)); // small file, fits in one chunk

        $fileSize = filesize($tmpFile);
        $round1 = [
            "file_exist"    => false,
            "upload_key"    => "uk_abc",
            "ex"            => "",
            "channel_list"  => [["id" => 1, "offset" => 0, "len" => $fileSize]],
            "upload_state"  => 1,
        ];
        $round2 = [
            "upload_state" => 2,
            "file_id"      => "new_file_456",
            "filename"     => basename($tmpFile),
        ];

        $client = new MockClient([$round1, $round2]);

        try {
            $res = $client->upload($tmpFile);
            $this->assertEquals("new_file_456", $res["file_id"]);
        } finally {
            unlink($tmpFile);
        }
    }
}
