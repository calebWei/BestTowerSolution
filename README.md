# BestTowerSolution
- Given the farm ID, find the tower in the farm with the highest average RSSI
- Decided to try Golang for the first time

## Instruction to run
Download `main.exe` from the latest release, and run `./main.exe [farm_id]`
![image](https://github.com/calebWei/BestTowerSolution/assets/100410646/45dbf30a-8dc8-474a-813b-468c5ccb9032)

## Attempts to resolve `Access Denied` Issue
- Trying to spot any problem with the URL itself
- Setting requested content type to `text/csv`
- Messing with User agent attribute in header
- Capturing packets with Wireshark (Don't think the issue is apparent from this view)

Current hypothesis:
- The object is set to private access
- The presigned URLs are missing appropriate credentials
- The presigned URLs already expired (?)
