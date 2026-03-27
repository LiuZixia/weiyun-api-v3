import unittest
import argparse
import sys
import os
import time

from weiyun_api import WeiyunClient

# Real integration tests hitting the Weiyun API endpoint.
# Requires WEIYUN_MCP_TOKEN environment variable.

class TestRealIntegration(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        token = os.environ.get("WEIYUN_MCP_TOKEN")
        if not token:
            print("Error: WEIYUN_MCP_TOKEN is not set.")
            sys.exit(1)
        cls.client = WeiyunClient(token=token)
        cls.ci_dir_key = cls._get_ci_dir()

    @classmethod
    def _get_root_pdir(cls):
        res = cls.client.list_files(limit=1)
        return res.get("pdir_key", "")

    @classmethod
    def _get_ci_dir(cls):
        """Finds the /CI directory in the root. Raises an error if not found."""
        offset = 0
        limit = 50
        while True:
            res = cls.client.list_files(limit=limit, offset=offset)
            dir_list = res.get("dir_list", [])
            for d in dir_list:
                if d.get("dir_name") == "CI":
                    print(f"✅ Found /CI directory: {d['dir_key']}")
                    return d["dir_key"]
            if res.get("finish_flag", True):
                break
            offset += limit
            
        raise RuntimeError("⚠️ /CI directory not found. Please create a folder named 'CI' in the root of your Weiyun drive for integration tests to work.")

    def test_list_ci_dir(self):
        res = self.client.list_files(limit=10, dir_key=self.ci_dir_key, pdir_key=self.__class__._get_root_pdir())
        self.assertIn("file_list", res)
        self.assertIn("dir_list", res)
        print("✅ Successfully listed files in /CI directory.")

    def test_upload_file(self):
        filename = f"test_python_upload_{int(time.time())}"
        with open(filename, "w") as f:
            f.write("Hello from CI Upload Test!")
        
        try:
            res = self.client.upload(file_path=filename, pdir_key=self.ci_dir_key)
            self.assertIn("file_id", res)
            print(f"✅ Successfully uploaded {filename} to /CI.")
        finally:
            if os.path.exists(filename):
                os.remove(filename)

    def test_download_file(self):
        # We must ensure tencent-weiyun.zip exists in /CI, or upload it first
        filename = "tencent-weiyun.zip"
        with open(filename, "w") as f:
            f.write("Dummy ZIP content for download testing")
            
        up_res = self.client.upload(file_path=filename, pdir_key=self.ci_dir_key)
        file_id = up_res["file_id"]
        
        # Download
        dl_res = self.client.download([{"file_id": file_id, "pdir_key": self.ci_dir_key}])
        self.assertIn("items", dl_res)
        self.assertGreater(len(dl_res["items"]), 0)
        self.assertIn("https_download_url", dl_res["items"][0])
        print(f"✅ Successfully obtained real download link for {filename}.")
        
        if os.path.exists(filename):
            os.remove(filename)

    def test_delete_file(self):
        filename = f"test_python_delete_{int(time.time())}.txt"
        with open(filename, "w") as f:
            f.write("Will automatically delete me!")
            
        try:
            up_res = self.client.upload(file_path=filename, pdir_key=self.ci_dir_key)
            file_id = up_res["file_id"]
            
            # Delete file using weiyun.delete
            del_res = self.client.delete(file_list=[{"file_id": file_id, "pdir_key": self.ci_dir_key}], delete_completely=True)
            self.assertIn("freed_index_cnt", del_res)
            self.assertGreaterEqual(del_res["freed_index_cnt"], 1)
            print(f"✅ Successfully deleted {filename} from /CI.")
        finally:
            if os.path.exists(filename):
                os.remove(filename)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run specific integration test")
    parser.add_argument("--test", choices=["list", "upload", "download", "delete", "all"], default="all")
    args = parser.parse_args()

    suite = unittest.TestSuite()
    if args.test == "list" or args.test == "all":
        suite.addTest(TestRealIntegration('test_list_ci_dir'))
    if args.test == "upload" or args.test == "all":
        suite.addTest(TestRealIntegration('test_upload_file'))
    if args.test == "download" or args.test == "all":
        suite.addTest(TestRealIntegration('test_download_file'))
    if args.test == "delete" or args.test == "all":
        suite.addTest(TestRealIntegration('test_delete_file'))

    runner = unittest.TextTestRunner(verbosity=2)
    result = runner.run(suite)
    sys.exit(not result.wasSuccessful())
