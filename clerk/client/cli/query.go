package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	clerkTypes "github.com/maticnetwork/heimdall/clerk/types"
	hmClient "github.com/maticnetwork/heimdall/client"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	// Group supply queries under a subcommand
	queryCmds := &cobra.Command{
		Use:                        clerkTypes.ModuleName,
		Short:                      "Querying commands for the clerk module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       hmClient.ValidateCmd,
	}

	// clerk query command
	queryCmds.AddCommand(
		client.GetCommands(
			GetStateRecord(cdc),
		)...,
	)

	return queryCmds
}

// GetStateRecord get state record
func GetStateRecord(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "show state record",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			recordIDStr := viper.GetString(FlagRecordID)
			if recordIDStr == "" {
				return fmt.Errorf("record id cannot be empty")
			}

			recordID, err := strconv.ParseUint(recordIDStr, 10, 64)
			if err != nil {
				return err
			}

			// get query params
			queryParams, err := cliCtx.Codec.MarshalJSON(clerkTypes.NewQueryRecordParams(recordID))
			if err != nil {
				return err
			}

			// fetch state reocrd
			res, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/%s", clerkTypes.QuerierRoute, clerkTypes.QueryRecord),
				queryParams,
			)

			if err != nil {
				return err
			}

			if len(res) == 0 {
				return errors.New("Record not found")
			}

			var _record clerkTypes.EventRecord
			if err := cdc.UnmarshalJSON(res, &_record); err != nil {
				fmt.Printf("Unable to unmarshall Error: %v", err)
				return err
			}

			fmt.Println(_record.String())
			return nil
		},
	}

	cmd.Flags().Uint64(FlagRecordID, 0, "--id=<record ID here>")
	cmd.MarkFlagRequired(FlagRecordID)

	return cmd
}