SOFTWARE_UPGRADE_NAME="v11"
UPGRADE_HEIGHT=22826069

UPGRADE_INFO=$(jq -n '
  {
      "binaries": {
          "linux/amd64": "https://github.com/classic-terra/core/releases/download/v3.4.0-rc.0/terra_3.4.0-rc.0_Linux_x86_64.tar.gz?checksum=sha256:2402fca560086e35693eb7dfd3f1e7a0923475c4568719ab2f0bcf7ee91b4f3b",
      }
  }')



terrad tx gov submit-legacy-proposal software-upgrade "$SOFTWARE_UPGRADE_NAME" --upgrade-height $UPGRADE_HEIGHT --upgrade-info "$UPGRADE_INFO" --title "Upgrade to v11" --description "Orbit: Upgrade regarding the first part of the unforking proposal"  --from orbit-testnet --keyring-backend file --chain-id "rebel-2" --gas-prices 30uluna --gas 3000000 --node https://rpc.luncblaze.com:443