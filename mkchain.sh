mkdir -p chain/
./hivechain generate \
    -outdir chain/       \
    -length 20           \
    -tx-count 4          \
    -tx-interval 1       \
    -fork-interval 0     \
    -lastfork prague     \
    -outputs genesis,chain,forkenv,headstate,txinfo,accounts,headfcu
