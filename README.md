# GoRedLoot

A tool to collect secrets (keys and passwords) and stage (compress and encrypt) them for exfiltration.
More details: https://lockboxx.blogspot.com/2018/06/goredloot.html

# Usage

- The tool takes two command line arguments when invoked, the directory to recursively search and the output file to create. 

-- Example: ./GoRedLoot [directory to recursivly search] [out file]

- The tool has five primary, hardcoded, internal configuration options. 

-- The first, and one you defiantly want to change, is the encryption password. 

-- The next four are essentially your search criteria, and they are ignoreFiles, includeFiles, ignoreContents, and includeContents, and are processed in that order. 

- Its also important to understand the double zipping process that occurs on the output file: 

-- The first zip wrapper retains all of the collected files meta-information, such as the file names and file properties. 

-- The second zip wrapper strips all of this information and encrypts the zip archive with the hard coded password.
