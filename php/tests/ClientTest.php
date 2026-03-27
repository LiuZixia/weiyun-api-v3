<?php
require_once __DIR__ . '/../src/Weiyun/Client.php';

use PHPUnit\Framework\TestCase;
use Weiyun\Client;
use Weiyun\SHA1;

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
}
