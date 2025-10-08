package rpc

import (
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

// Request represents an RPC request message containing a payload and one or more signatures.
// The Request structure supports multi-signature scenarios where multiple parties need to
// authorize an operation. Signatures are typically created by signing the marshaled payload.
//
// The JSON representation of a Request is:
//
//	{
//	  "req": [requestId, method, params, timestamp],
//	  "sig": [signature1, signature2, ...]
//	}
//
// Where the "req" field contains the compactly encoded payload and "sig" contains
// an array of signatures authorizing the request.
type Request struct {
	// Req contains the request payload with method, parameters, and metadata
	Req Payload `json:"req"`
	// Sig contains one or more signatures authorizing this request.
	// Multiple signatures enable multi-sig authorization scenarios.
	Sig []sign.Signature `json:"sig"`
}

// NewRequest creates a new Request with the given payload and optional signatures.
// If no signatures are provided, the Sig field will be an empty slice, which
// typically means signatures will be added later by the transport layer.
//
// Example usage:
//
//	// Create request without signatures (to be signed later)
//	request := NewRequest(payload)
//
//	// Create request with single signature
//	request := NewRequest(payload, signature)
//
//	// Create request with multiple signatures for multi-sig auth
//	request := NewRequest(payload, sig1, sig2, sig3)
func NewRequest(payload Payload, sig ...sign.Signature) Request {
	return Request{
		Req: payload,
		Sig: sig,
	}
}

// GetSigners recovers and returns the addresses of all signers from the request signatures.
// This method is used to identify which addresses authorized this request by recovering
// the public key addresses from the cryptographic signatures.
//
// Returns:
//   - A slice of addresses, one for each signature in the request
//   - An error if signature recovery fails for any signature
//
// Example usage:
//
//	// Verify request signers
//	addresses, err := request.GetSigners()
//	if err != nil {
//	    return rpc.Errorf("invalid signatures: %v", err)
//	}
//
//	// Check if a specific address signed the request
//	for _, addr := range addresses {
//	    if addr == authorizedAddress {
//	        // Process the authorized request
//	    }
//	}
func (r Request) GetSigners() ([]sign.Address, error) {
	return recoverPayloadSigners(r.Req, r.Sig)
}

// Response represents an RPC response message containing a payload and one or more signatures.
// The structure mirrors Request, allowing responses to be cryptographically signed by the
// server or multiple parties in a distributed system.
//
// The JSON representation of a Response is:
//
//	{
//	  "res": [requestId, method, params, timestamp],
//	  "sig": [signature1, signature2, ...]
//	}
//
// The response payload typically contains the same RequestID as the original request,
// allowing clients to match responses to their requests.
type Response struct {
	// Res contains the response payload with results or error information
	Res Payload `json:"res"`
	// Sig contains one or more signatures authenticating this response.
	// This ensures clients can verify the response came from authorized servers.
	Sig []sign.Signature `json:"sig"`
}

// NewResponse creates a new Response with the given payload and optional signatures.
// The payload should contain the same RequestID as the original request to enable
// request-response matching on the client side.
//
// Example usage:
//
//	// Create response for a successful operation
//	resultParams, _ := NewParams(map[string]interface{}{"status": "success", "txHash": "0xabc"})
//	responsePayload := NewPayload(request.Req.RequestID, request.Req.Method, resultParams)
//	response := NewResponse(responsePayload, serverSignature)
//
//	// Create error response
//	errorParams, _ := NewParams(map[string]interface{}{"error": "insufficient funds"})
//	errorPayload := NewPayload(request.Req.RequestID, request.Req.Method, errorParams)
//	response := NewResponse(errorPayload, serverSignature)
func NewResponse(payload Payload, sig ...sign.Signature) Response {
	return Response{
		Res: payload,
		Sig: sig,
	}
}

// GetSigners recovers and returns the addresses of all signers from the response signatures.
// This method allows clients to verify that responses came from authorized servers
// by recovering the public key addresses from the cryptographic signatures.
//
// Returns:
//   - A slice of addresses, one for each signature in the response
//   - An error if signature recovery fails for any signature
//
// Example usage:
//
//	// Verify response came from trusted server
//	addresses, err := response.GetSigners()
//	if err != nil {
//	    return fmt.Errorf("invalid response signatures: %w", err)
//	}
//
//	if !contains(addresses, trustedServerAddress) {
//	    return fmt.Errorf("response not from trusted server")
//	}
func (r Response) GetSigners() ([]sign.Address, error) {
	return recoverPayloadSigners(r.Res, r.Sig)
}

// NewErrorResponse creates a Response containing an error message.
// This is a convenience function that combines error parameter creation
// and response construction in a single call.
//
// Parameters:
//   - requestID: The ID from the original request
//   - errMsg: The error message to send to the client
//   - sig: Optional signatures to authenticate the error response
//
// Example usage:
//
//	// In an RPC handler when an error occurs
//	if err := validateRequest(request); err != nil {
//	    return NewErrorResponse(request.Req.RequestID, err.Error(), serverSignature)
//	}
//
//	// Creating an error response without signatures
//	errorResponse := NewErrorResponse(12345, "insufficient balance")
//
// The resulting response will have params in the format: {"error": "<errMsg>"}
func NewErrorResponse(requestID uint64, errMsg string, sig ...sign.Signature) Response {
	errParams := NewErrorParams(errMsg)
	errPayload := NewPayload(requestID, ErrorMethod.String(), errParams)
	return NewResponse(errPayload, sig...)
}

// Error checks if the Response contains an error and returns it.
// This method extracts any error stored in the response payload's params
// under the standard "error" key.
//
// Returns:
//   - An error if the response contains an error message
//   - nil if the response represents a successful operation
//
// This is typically used by clients to check if an RPC call failed:
//
//	// After receiving and unmarshaling a response
//	var response Response
//	if err := json.Unmarshal(data, &response); err != nil {
//	    return fmt.Errorf("failed to unmarshal response: %w", err)
//	}
//
//	// Check if the response contains an error
//	if err := response.Error(); err != nil {
//	    return fmt.Errorf("RPC call failed: %w", err)
//	}
//
//	// Process successful response
//	var result TransferResult
//	response.Res.Params.Translate(&result)
//
// This method is designed to work with error responses created by NewErrorResponse
// or any response where errors are stored using NewErrorParams.
func (r Response) Error() error {
	if r.Res.Method != ErrorMethod.String() {
		return nil
	}

	return r.Res.Params.Error()
}

// recoverPayloadSigners is an internal helper function that recovers signer addresses
// from a payload and its associated signatures. This function is used by both
// Request.GetSigners and Response.GetSigners to perform the actual signature recovery.
//
// The function works by:
// 1. Computing the hash of the payload
// 2. For each signature, creating an address recoverer
// 3. Recovering the address that created each signature
//
// Parameters:
//   - payload: The payload that was signed
//   - sigs: The signatures to recover addresses from
//
// Returns:
//   - A slice of recovered addresses in the same order as the signatures
//   - An error if hash computation or any signature recovery fails
//
// This function is critical for the security of the protocol as it enables
// verification of message authenticity and authorization.
func recoverPayloadSigners(payload Payload, sigs []sign.Signature) ([]sign.Address, error) {
	payloadHash, err := payload.Hash()
	if err != nil {
		return nil, err
	}

	var addrs []sign.Address
	for _, s := range sigs {
		recoverer, err := sign.NewAddressRecovererFromSignature(s)
		if err != nil {
			return nil, err
		}

		addr, err := recoverer.RecoverAddress(payloadHash, s)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}

	return addrs, nil
}
