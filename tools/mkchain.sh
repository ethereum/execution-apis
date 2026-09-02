mkdir -p chain/
./hivechain generate \
    -outdir chain/       \
    -length 54           \
    -tx-count 4          \
    -tx-interval 1       \
    -fork-interval 3     \
    -lastfork bpo2       \
    -outputs genesis,chain,forkenv,headstate,txinfo,accounts,headfcu
