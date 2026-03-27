<?php
/**
 * Weiyun PHP Integration Tests
 * Mirrors python/integration_tests.py
 *
 * Usage:
 *   WEIYUN_MCP_TOKEN=xxx php integration_tests.php [--test list|upload|download|delete|all]
 *
 * Requires a folder named "CI" at the root of your Weiyun drive.
 */

require_once __DIR__ . '/src/Weiyun/Client.php';
use Weiyun\Client;

// ─── Bootstrap ───────────────────────────────────────────────────────────────

$token = getenv('WEIYUN_MCP_TOKEN');
if (!$token) {
    fwrite(STDERR, "Error: WEIYUN_MCP_TOKEN is not set.\n");
    exit(1);
}

$mcpUrl = getenv('WEIYUN_MCP_URL') ?: 'https://www.weiyun.com/api/v3/mcpserver';
$client = new Client($token, $mcpUrl);

// ─── Helpers ─────────────────────────────────────────────────────────────────

function getRootPdir(Client $client): string {
    $res = $client->listFiles(1);
    return $res['pdir_key'] ?? '';
}

function getCiDirKey(Client $client): string {
    $offset = 0;
    $limit  = 50;
    while (true) {
        $res     = $client->listFiles($limit, 0, $offset);
        $dirList = $res['dir_list'] ?? [];
        foreach ($dirList as $d) {
            if (($d['dir_name'] ?? '') === 'CI') {
                echo "✅ Found /CI directory: {$d['dir_key']}\n";
                return $d['dir_key'];
            }
        }
        if ($res['finish_flag'] ?? true) {
            break;
        }
        $offset += $limit;
    }
    throw new RuntimeException(
        "⚠️  /CI directory not found. Please create a folder named 'CI' " .
        "in the root of your Weiyun drive for integration tests to work."
    );
}

// ─── Tests ───────────────────────────────────────────────────────────────────

$passed = 0;
$failed = 0;

function run_test(string $name, callable $fn): void {
    global $passed, $failed;
    echo "\n▶ $name\n";
    try {
        $fn();
        $passed++;
    } catch (Throwable $e) {
        echo "❌ FAILED: " . $e->getMessage() . "\n";
        $failed++;
    }
}

function assert_has_key(array $arr, string $key): void {
    if (!array_key_exists($key, $arr)) {
        throw new RuntimeException("Expected key '{$key}' not found in response.");
    }
}

// Locate the /CI directory once
$ciDirKey  = getCiDirKey($client);
$rootPdir  = getRootPdir($client);

// ── test_list_ci_dir ─────────────────────────────────────────────────────────
function test_list_ci_dir(Client $client, string $ciDirKey, string $rootPdir): void {
    $res = $client->listFiles(10, 0, 0, $ciDirKey, $rootPdir);
    assert_has_key($res, 'file_list');
    assert_has_key($res, 'dir_list');
    echo "✅ Successfully listed files in /CI directory.\n";
}

// ── test_upload_file ─────────────────────────────────────────────────────────
function test_upload_file(Client $client, string $ciDirKey): void {
    $filename = "test_php_upload_" . time();
    file_put_contents($filename, "Hello from CI Upload Test!");
    try {
        $res = $client->upload($filename, $ciDirKey);
        assert_has_key($res, 'file_id');
        echo "✅ Successfully uploaded {$filename} to /CI.\n";
    } finally {
        if (file_exists($filename)) unlink($filename);
    }
}

// ── test_download_file ───────────────────────────────────────────────────────
function test_download_file(Client $client, string $ciDirKey): void {
    $filename = "tencent-weiyun.zip";
    file_put_contents($filename, "Dummy ZIP content for download testing");
    $upRes  = $client->upload($filename, $ciDirKey);
    $fileId = $upRes['file_id'];

    $dlRes = $client->download([["file_id" => $fileId, "pdir_key" => $ciDirKey]]);
    assert_has_key($dlRes, 'items');
    if (count($dlRes['items']) === 0) {
        throw new RuntimeException("Download items list is empty.");
    }
    assert_has_key($dlRes['items'][0], 'https_download_url');
    echo "✅ Successfully obtained real download link for {$filename}.\n";

    if (file_exists($filename)) unlink($filename);
}

// ── test_delete_file ─────────────────────────────────────────────────────────
function test_delete_file(Client $client, string $ciDirKey): void {
    $filename = "test_php_delete_" . time() . ".txt";
    file_put_contents($filename, "Will automatically delete me!");
    try {
        $upRes  = $client->upload($filename, $ciDirKey);
        $fileId = $upRes['file_id'];

        $delRes = $client->delete(
            [["file_id" => $fileId, "pdir_key" => $ciDirKey]],
            null,
            true
        );
        assert_has_key($delRes, 'freed_index_cnt');
        if ($delRes['freed_index_cnt'] < 1) {
            throw new RuntimeException("freed_index_cnt should be >= 1.");
        }
        echo "✅ Successfully deleted {$filename} from /CI.\n";
    } finally {
        if (file_exists($filename)) unlink($filename);
    }
}

// ─── Argument Parsing & Runner ───────────────────────────────────────────────

$opts  = getopt('', ['test:']);
$which = $opts['test'] ?? 'all';

if (in_array($which, ['list',     'all'])) run_test('test_list_ci_dir',
    fn() => test_list_ci_dir($client, $ciDirKey, $rootPdir));
if (in_array($which, ['upload',   'all'])) run_test('test_upload_file',
    fn() => test_upload_file($client, $ciDirKey));
if (in_array($which, ['download', 'all'])) run_test('test_download_file',
    fn() => test_download_file($client, $ciDirKey));
if (in_array($which, ['delete',   'all'])) run_test('test_delete_file',
    fn() => test_delete_file($client, $ciDirKey));

echo "\n────────────────────────────────\n";
echo "Results: {$passed} passed, {$failed} failed\n";
exit($failed > 0 ? 1 : 0);
