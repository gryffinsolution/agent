#!/usr/bin/perl
use strict;

my $pid=$ARGV[0];

while (1) {
	system("ps --no-heading -o etime,pcpu,pmem,rss,vsz -p $pid |tee -a res.csv");
	sleep 1;
}
