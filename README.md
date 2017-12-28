USAGE
##build the project

##启动本机的玩客币钱包
###./auto -h
##./auto -dataDir=-dataDir=/Users/yourhome/Library/OTCWalletData


##curl http://127.0.0.1:65399/feedmonkeys -X POST --data '{"from_address":"0x8a47675ceb86c2593041729","to_address":"0xfd6878db561f27feb46d436afab8c1bcc997a132","pwd":"youpassword","monkeys":[{"id":"55414","limit":"4.37","mode":"min"},{"id":"55414","limit":"5","mode":"min"}]}'

##喂养模式 min为每次喂最小量 最多次喂养 每次喂养0.xxx
        max为只喂养一次，喂最大值
##绕过玩客币钱包直接调用geth进行转账， 封号危险未知
##程序默认喂饱一只猴会向作者地址转账0.5WKC， 可以选择注释掉对应代码~
##勿用大额钱包操作， 翻车后果自负自负。。。。。。


