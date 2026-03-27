import unittest
from unittest.mock import patch, MagicMock
from weiyun_api import WeiyunClient

class TestWeiyunClient(unittest.TestCase):
    def setUp(self):
        self.client = WeiyunClient(token="test_token")

    @patch("weiyun_api.requests.post")
    def test_list_files(self, mock_post):
        mock_resp = MagicMock()
        mock_resp.json.return_value = {
            "result": {
                "content": [
                    {
                        "type": "text",
                        "text": "{\"file_list\": [{\"file_id\": \"123\"}], \"dir_list\": []}"
                    }
                ]
            }
        }
        mock_post.return_value = mock_resp
        
        res = self.client.list_files(limit=10)
        self.assertIn("file_list", res)
        self.assertEqual(res["file_list"][0]["file_id"], "123")

    @patch("weiyun_api.requests.post")
    def test_download(self, mock_post):
        mock_resp = MagicMock()
        mock_resp.json.return_value = {
            "result": {
                "content": [
                    {
                        "type": "text",
                        "text": "{\"items\": [{\"file_id\": \"123\", \"https_download_url\": \"url\"}]}"
                    }
                ]
            }
        }
        mock_post.return_value = mock_resp
        
        res = self.client.download([{"file_id": "123", "pdir_key": "abc"}])
        self.assertIn("items", res)

if __name__ == "__main__":
    unittest.main()
