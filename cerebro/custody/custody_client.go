package custody

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

const (
	minCustodyChallengePeriod = 3600 // Default challenge period in seconds
)

type CustodyClient struct {
}

func NewCustodyClient() *CustodyClient {
	return &CustodyClient{}
}

func (c *CustodyClient) OpenChannel(
	wallet, signer sign.Signer,
	chainID uint32, chainRPC string,
	custodyAddress, adjudicatorAddress, brokerAddress, tokenAddress common.Address,
	challenge, nonce uint64,
	brokerSig sign.Signature,
) (string, error) {
	client, err := ethclient.Dial(chainRPC)
	if err != nil {
		return "", err
	}

	custody, err := NewCustody(custodyAddress, client)
	if err != nil {
		return "", err
	}

	if challenge == 0 {
		challenge = minCustodyChallengePeriod
	} else if challenge < minCustodyChallengePeriod {
		return "", fmt.Errorf("challenge period must be at least %d seconds", minCustodyChallengePeriod)
	}

	walletAddress := common.HexToAddress(wallet.PublicKey().Address().String())
	signerAddress := common.HexToAddress(signer.PublicKey().Address().String())
	channel := Channel{
		Adjudicator:  adjudicatorAddress,
		Participants: []common.Address{signerAddress, brokerAddress},
		Challenge:    challenge,
		Nonce:        nonce,
	}
	initial := State{
		Intent:  1, // IntentINITIALIZE
		Version: new(big.Int).SetUint64(0),
		Data:    []byte(""),
		Allocations: []Allocation{
			{
				Destination: walletAddress,
				Token:       tokenAddress,
				Amount:      big.NewInt(0), // Initial amount is zero
			},
			{
				Destination: brokerAddress,
				Token:       tokenAddress,
				Amount:      big.NewInt(0), // Initial amount is zero
			},
		},
	}

	channelID, err := GetChannelID(channel, chainID)
	if err != nil {
		return "", fmt.Errorf("failed to compute channel ID: %w", err)
	}

	initialStateData, err := EncodeState(channelID, initial)
	if err != nil {
		return "", fmt.Errorf("failed to encode initial state: %w", err)
	}

	dataHash := crypto.Keccak256Hash(initialStateData)
	userSig, err := signer.Sign(dataHash.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to sign initial state: %w", err)
	}
	initial.Sigs = [][]byte{userSig, brokerSig}

	txOpts := signerTxOpts(wallet, chainID)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to suggest gas price: %w", err)
	}
	txOpts.GasPrice = gasPrice.Add(gasPrice, gasPrice)

	tx, err := custody.Create(txOpts, channel, initial)
	if err != nil {
		return "", fmt.Errorf("failed to create custody channel: %w", err)
	}

	if _, err := bind.WaitMined(context.Background(), client, tx.Hash()); err != nil {
		return "", err
	}

	fmt.Printf("Channel created successfully: %s\n", tx.Hash().Hex())
	return channelID.Hex(), nil
}

func (c *CustodyClient) CloseChannel(
	wallet, signer sign.Signer,
	chainID uint32, chainRPC string,
	custodyAddress common.Address,
	channelID common.Hash,
	version *big.Int,
	allocations []Allocation,
	brokerSig sign.Signature,
) error {
	client, err := ethclient.Dial(chainRPC)
	if err != nil {
		return err
	}

	custody, err := NewCustody(custodyAddress, client)
	if err != nil {
		return err
	}

	state := State{
		Intent:      3, // IntentFINALIZE
		Version:     version,
		Data:        []byte(""),
		Allocations: allocations,
	}
	stateData, err := EncodeState(channelID, state)
	if err != nil {
		return fmt.Errorf("failed to encode initial state: %w", err)
	}

	dataHash := crypto.Keccak256Hash(stateData)
	userSig, err := signer.Sign(dataHash.Bytes())
	if err != nil {
		return fmt.Errorf("failed to sign initial state: %w", err)
	}
	state.Sigs = [][]byte{userSig, brokerSig}

	txOpts := signerTxOpts(wallet, chainID)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %w", err)
	}
	txOpts.GasPrice = gasPrice.Add(gasPrice, gasPrice)

	tx, err := custody.Close(txOpts, channelID, state, nil)
	if err != nil {
		return fmt.Errorf("failed to close custody channel: %w", err)
	}

	if _, err := bind.WaitMined(context.Background(), client, tx.Hash()); err != nil {
		return err
	}

	fmt.Printf("Channel closed successfully: %s\n", tx.Hash().Hex())
	return nil
}

func (c *CustodyClient) Resize(
	wallet, signer sign.Signer,
	chainID uint32, chainRPC string,
	custodyAddress common.Address,
	channelID common.Hash,
	version *big.Int,
	data []byte,
	allocations []Allocation,
	brokerSig sign.Signature,
) error {
	client, err := ethclient.Dial(chainRPC)
	if err != nil {
		return err
	}

	custody, err := NewCustody(custodyAddress, client)
	if err != nil {
		return err
	}

	channelData, err := custody.GetChannelData(&bind.CallOpts{}, channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel data: %w", err)
	}

	state := State{
		Intent:      2, // IntentRESIZE
		Version:     version,
		Data:        data,
		Allocations: allocations,
	}
	stateData, err := EncodeState(channelID, state)
	if err != nil {
		return fmt.Errorf("failed to encode initial state: %w", err)
	}

	dataHash := crypto.Keccak256Hash(stateData)
	userSig, err := signer.Sign(dataHash.Bytes())
	if err != nil {
		return fmt.Errorf("failed to sign initial state: %w", err)
	}
	state.Sigs = [][]byte{userSig, brokerSig}

	txOpts := signerTxOpts(wallet, chainID)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %w", err)
	}
	txOpts.GasPrice = gasPrice.Add(gasPrice, gasPrice)

	tx, err := custody.Resize(txOpts, channelID, state, []State{channelData.LastValidState})
	if err != nil {
		return fmt.Errorf("failed to resize custody channel: %w", err)
	}

	if _, err := bind.WaitMined(context.Background(), client, tx.Hash()); err != nil {
		return err
	}

	fmt.Printf("Channel resized successfully: %s\n", tx.Hash().Hex())
	return nil
}

func (c *CustodyClient) Deposit(
	wallet sign.Signer,
	chainID uint32, chainRPC string,
	custodyAddress, tokenAddress common.Address,
	amount *big.Int,
) error {
	client, err := ethclient.Dial(chainRPC)
	if err != nil {
		return err
	}

	custody, err := NewCustody(custodyAddress, client)
	if err != nil {
		return err
	}

	txOpts := signerTxOpts(wallet, chainID)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %w", err)
	}
	txOpts.GasPrice = gasPrice.Add(gasPrice, gasPrice)

	walletAddress := common.HexToAddress(wallet.PublicKey().Address().String())
	tx, err := custody.Deposit(txOpts, walletAddress, tokenAddress, amount)
	if err != nil {
		return fmt.Errorf("failed to deposit into custody: %w", err)
	}

	if _, err := bind.WaitMined(context.Background(), client, tx.Hash()); err != nil {
		return err
	}

	return nil
}

func (c *CustodyClient) Withdraw(
	wallet sign.Signer,
	chainID uint32, chainRPC string,
	custodyAddress, tokenAddress common.Address,
	amount *big.Int,
) error {
	client, err := ethclient.Dial(chainRPC)
	if err != nil {
		return err
	}

	custody, err := NewCustody(custodyAddress, client)
	if err != nil {
		return err
	}

	txOpts := signerTxOpts(wallet, chainID)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %w", err)
	}
	txOpts.GasPrice = gasPrice.Add(gasPrice, gasPrice)

	tx, err := custody.Withdraw(txOpts, tokenAddress, amount)
	if err != nil {
		return fmt.Errorf("failed to deposit into custody: %w", err)
	}

	if _, err := bind.WaitMined(context.Background(), client, tx.Hash()); err != nil {
		return err
	}

	return nil
}

func (c *CustodyClient) GetLedgerBalance(
	chainID uint32, chainRPC string,
	custodyAddress, walletAddress, tokenAddress common.Address,
) (*big.Int, error) {
	client, err := ethclient.Dial(chainRPC)
	if err != nil {
		return nil, err
	}

	custody, err := NewCustody(custodyAddress, client)
	if err != nil {
		return nil, err
	}

	balances, err := custody.GetAccountsBalances(
		&bind.CallOpts{},
		[]common.Address{walletAddress},
		[]common.Address{tokenAddress},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances: %w", err)
	}
	if len(balances) == 0 || len(balances[0]) == 0 {
		return nil, fmt.Errorf("no balances found for wallet %s on custody %s", walletAddress.Hex(), custodyAddress.Hex())
	}

	return balances[0][0], nil
}

func (c *CustodyClient) GetChannelBalance(
	chainID uint32, chainRPC string,
	custodyAddress common.Address,
	channelID common.Hash,
	tokenAddress common.Address,
) (*big.Int, error) {
	client, err := ethclient.Dial(chainRPC)
	if err != nil {
		return nil, err
	}

	custody, err := NewCustody(custodyAddress, client)
	if err != nil {
		return nil, err
	}

	balances, err := custody.GetChannelBalances(
		&bind.CallOpts{},
		channelID,
		[]common.Address{tokenAddress},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances: %w", err)
	}
	if len(balances) == 0 {
		return nil, fmt.Errorf("no balances found for channel %s on custody %s", channelID.Hex(), custodyAddress.Hex())
	}

	return balances[0], nil
}

// EncodeState encodes channel state into a byte array using channelID, intent, version, state data, and allocations.
func EncodeState(channelID common.Hash, s State) ([]byte, error) {
	allocationType, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
		{Name: "destination", Type: "address"},
		{Name: "token", Type: "address"},
		{Name: "amount", Type: "uint256"},
	})
	if err != nil {
		return nil, err
	}

	intentType, err := abi.NewType("uint8", "", nil)
	if err != nil {
		return nil, err
	}
	versionType, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // channelID
		{Type: intentType},               // intent
		{Type: versionType},              // version
		{Type: abi.Type{T: abi.BytesTy}}, // stateData
		{Type: allocationType},           // allocations (tuple[])
	}

	packed, err := args.Pack(channelID, s.Intent, s.Version, s.Data, s.Allocations)
	if err != nil {
		return nil, err
	}
	return packed, nil
}

// GetChannelID computes the channel ID
func GetChannelID(ch Channel, chainID uint32) (common.Hash, error) {
	participantsType, err := abi.NewType("address[]", "", nil)
	if err != nil {
		return common.Hash{}, err
	}
	uint64Type, err := abi.NewType("uint64", "", nil)
	if err != nil {
		return common.Hash{}, err
	}
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return common.Hash{}, err
	}

	args := abi.Arguments{
		{Type: participantsType},           // participants
		{Type: abi.Type{T: abi.AddressTy}}, // adjudicator
		{Type: uint64Type},                 // challenge
		{Type: uint64Type},                 // nonce
		{Type: uint256Type},                // chainID
	}

	bigChainID := new(big.Int).SetUint64(uint64(chainID))
	packed, err := args.Pack(ch.Participants, ch.Adjudicator, ch.Challenge, ch.Nonce, bigChainID)
	if err != nil {
		return common.Hash{}, err
	}

	return crypto.Keccak256Hash(packed), nil
}

func signerTxOpts(signer sign.Signer, chainID uint32) *bind.TransactOpts {
	bigChainID := big.NewInt(int64(chainID))
	signingMethod := types.LatestSignerForChainID(bigChainID)
	signerAddress := common.HexToAddress(signer.PublicKey().Address().String())
	signerFn := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != signerAddress {
			return nil, bind.ErrNotAuthorized
		}

		hash := signingMethod.Hash(tx).Bytes()
		sig, err := signer.Sign(hash)
		if err != nil {
			return nil, err
		}

		if sig[64] >= 27 {
			sig[64] -= 27
		}

		return tx.WithSignature(signingMethod, sig)
	}

	return &bind.TransactOpts{
		From:    signerAddress,
		Signer:  signerFn,
		Context: context.Background(),
	}
}
