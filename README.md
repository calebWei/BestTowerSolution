# BestTowerSolution
- Given the farm ID, find the tower in the farm with the highest average RSSI
- Decided to try Golang for the first time

## Instruction to run
Download `main.exe` from the latest release, and run `./main.exe [farm_id]`
![image](https://github.com/calebWei/BestTowerSolution/assets/100410646/fb287d84-b839-4782-9b4b-55643edc5843)


## Attempts to resolve the `Access Denied` Issue
- Trying to spot any problem with the URL itself
- Setting requested content type to `text/csv`
- Messing with User agent attribute in request header
- Capturing packets with Wireshark (Don't think the issue is apparent from this view)
- Examining difference in request/response headers

Current hypothesis:
- The object is not publicly accessible
- If the object doesn't exist, could return 403 Access Denied without s3:ListBucket permission
- IAM Policy & Bucket Policy issue
- The presigned URLs are missing appropriate credentials
- The presigned URLs already expired (?)
