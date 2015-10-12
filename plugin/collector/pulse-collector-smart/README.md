##Pulse S.M.A.R.T. Collector plugin

#Description

This plugin collect metrics from Intel's SSDs. Raw data interpretation is based on [State Drive DC S3700 Series specification](http://www.intel.com/content/dam/www/public/us/en/documents/product-specifications/ssd-dc-s3700-spec.pdf). Other disks may have different attributes or different raw data formats.

#Metrics

 - /intel
	 - /disk
		 - /[DEVICE NAME (ex. sda)
			 - /reallocatedsectors - number of retired blocks
				- /normalized - shows percent remaining of allowable grown defect count
			 - /poweronhours - cumulative power-on time in hours
				 - /normalized - always 100
			 - /powercyclecount - cumulative number of power cycle events
				 - /normalized - always 100
			 - /availablereservedspace - available reserved space
				 - /normalized -  undocumented
			 - /programfailcount - total count of program fails
				 - /normalized - percent remaining of allowable program fails
			 - /erasefailcount - total count of erase fails
				 - /percent remaining of allowable erase fails
			 - /unexpectedpowerloss - cumulative number of unclean shutdowns
				 - /normalized - always 100
			 - /powerlossprotectionfailure - last test result as microseconds to discharge capacitor
				 - /sincelast - minutes since last test
				 - /tests - lifetime number of tests
				 - /normalized - 1 on test failure, 11 if capacitor tested in excessive temperature, otherwise 100
			 - /satadownshifts - number of times SATA interface selected lower signaling rate due to error
				 - /normalized - always 100
			 - /e2eerrors - number of LBA tag mismatches in end-to-end data protection path
				 - /normalized - always 100
			 - /uncorrectableerrors - number of errors that could not be recovered using Error Correction Code
				 - /normalized - always 100
			 - /casetemperature - SSD case temperature in Celsius
				 - /min - minimal value
				 - /max - maximal value
				 - /overcounter - number of times sampled temperature exceeds drive max operating temperature spec.
				 - /normalized value - (100-temperature in Celsius)
			 - /unsafeshutdowns - cumulative number of unsafe shutdowns
				 - /normalized - always 100
			 - /internaltemperature - device internal temperature in Celsius. Reading from PCB.
				 - /normalized - (150 - temperature in Celsius) or 100 if temperature is less than 50.
			 - /pendingsectors - number of current unrecoverable read errors that will be re-allocated on next write.
				 - /normalized - always 100.
			 - /crcerrors" - total number of encountered SATA CRC errors.
				 - /normalized - always 100
			 - /hostwrites - total number of sectors written by the host system
				 - /normalized - always 100
			 - /timedworkload
				 - /mediawear - measures the wear seen by the SSD (since reset of the workload timer, see timedworkload/time), as a percentage of the maximum rated cycles.
					 - /normalized - always 100
				 - /readpercent - shows the percentage of I/O operations that are read operations (since reset of the workload timer, see timedworkload/time)
					 - /normalized - always 100
				 - /time - number of minutes since starting workload timer
					 - /normalized - always 100
			 - /reservedblocks - number of reserved blocks remaining
				 - /normalized - percentage of reserved space available
			 - /wearout - always 0
				 - /number of cycles the NAND media has undergone. Declines linearly from 100 to 1 as the average erase cycle count increases from 0 to the maximum rated cycles. Once it reaches 1 the number will not decrease, although it is likely that significant additional wear can be put on the device.
			 - /thermalthrottle" - percent throttle status
				 - /eventcount - number of times thermal throttle has activated. Preserved over power cycles.
				 - /normalized - always 100
			 - /totallba
				 - /written - total number of sectors written by the host system
					 - /normalized - always 100
				 - /read - total number of sectors read by the host system
					 - /normalized - always 100
