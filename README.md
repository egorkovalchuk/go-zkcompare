
# Zookeeper Data Comparison Utility
This project is a utility for comparing data between two Zookeeper instances and for searching data by keywords. The utility supports multithreading and allows excluding specific paths from analysis.

## Description
The utility performs the following tasks:
1. Connecting to two Zookeeper instances.
2. Recursive traversal of the specified path in Zookeeper.
3. Comparing data between the source and destination.
4. Searching data by keywords.
5. Writing results to a log file.

## Usage
### Command-line Flags
* **-s** : Source Zookeeper address (mandatory parameter).
* **-d** : Destination Zookeeper address (mandatory parameter).
* **-p** : Path in Zookeeper for comparison or search (default: /).
* **-e** : Excluding paths (comma-separated, default: password).
* **-f** : Search string (if specified, launches search mode).
* **-debug** : Enable debug mode (outputs additional information).
* **-h** : Display help.
* **-v** : Display the utility version.

Find string
```bash
./zkcompare -s zoo01:2181 -f login -p /ps/config
```

Compare 
```bash
./zkcompare -s zoo01:2181 -d zoo02:2181 -p /ps/config -e password,login
```