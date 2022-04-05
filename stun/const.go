// Copyright 2016 Cong Ding
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stun

// Default server address and client name.
const (
	DefaultServerAddr   = "stun.ekiga.net:3478"
	DefaultSoftwareName = "StunClient"
)

const (
	magicCookie = 0x2112A442
	fingerprint = 0x5354554e
)

// NATType is the type of NAT described by int.
type NATType int

// NAT behavior type
type BehaviorType int

type NATBehavior struct {
	MappingType   BehaviorType
	FilteringType BehaviorType
}

// NAT types.
const (
	NATError NATType = iota
	NATUnknown
	NATNone
	NATBlocked
	NATFull
	NATSymmetric
	NATRestricted
	NATPortRestricted
	SymmetricUDPFirewall

	// Deprecated spellings of these constants
	NATSymetric             = NATSymmetric
	NATSymetricUDPFirewall  = SymmetricUDPFirewall
	NATSymmetricUDPFirewall = SymmetricUDPFirewall
)

const (
	BehaviorTypeUnknown BehaviorType = iota
	BehaviorTypeEndpoint
	BehaviorTypeAddr
	BehaviorTypeAddrAndPort
)

var natStr map[NATType]string
var natBehaviorTypeStr map[BehaviorType]string
var natNormalTypeStr map[NATBehavior]string

func init() {
	natStr = map[NATType]string{
		NATError:             "Test failed",
		NATUnknown:           "Unexpected response from the STUN server",
		NATBlocked:           "UDP is blocked",
		NATFull:              "Full cone NAT",
		NATSymmetric:         "Symmetric NAT",
		NATRestricted:        "Restricted NAT",
		NATPortRestricted:    "Port restricted NAT",
		NATNone:              "Not behind a NAT",
		SymmetricUDPFirewall: "Symmetric UDP firewall",
	}

	natBehaviorTypeStr = map[BehaviorType]string{
		BehaviorTypeEndpoint:    "EndpointIndependent",
		BehaviorTypeAddr:        "AddressDependent",
		BehaviorTypeAddrAndPort: "AddressAndPortDependent",
	}

	// Defined in RFC 3489
	natNormalTypeStr = map[NATBehavior]string{
		NATBehavior{BehaviorTypeEndpoint, BehaviorTypeEndpoint}:       "Full cone NAT",
		NATBehavior{BehaviorTypeEndpoint, BehaviorTypeAddr}:           "Restricted cone NAT",
		NATBehavior{BehaviorTypeEndpoint, BehaviorTypeAddrAndPort}:    "Port Restricted cone NAT",
		NATBehavior{BehaviorTypeAddrAndPort, BehaviorTypeAddrAndPort}: "Symmetric NAT",
	}
}

func (nat NATType) String() string {
	if s, ok := natStr[nat]; ok {
		return s
	}
	return "Unknown"
}

func (natBhType BehaviorType) String() string {
	if s, ok := natBehaviorTypeStr[natBhType]; ok {
		return s
	}
	return "Unknown"
}

func (natBehavior NATBehavior) NormalType() string {
	if s, ok := natNormalTypeStr[natBehavior]; ok {
		return s
	}
	return "Undefined"
}

const (
	errorTryAlternate                 = 300
	errorBadRequest                   = 400
	errorUnauthorized                 = 401
	errorUnassigned402                = 402
	errorForbidden                    = 403
	errorUnknownAttribute             = 420
	errorAllocationMismatch           = 437
	errorStaleNonce                   = 438
	errorUnassigned439                = 439
	errorAddressFamilyNotSupported    = 440
	errorWrongCredentials             = 441
	errorUnsupportedTransportProtocol = 442
	errorPeerAddressFamilyMismatch    = 443
	errorConnectionAlreadyExists      = 446
	errorConnectionTimeoutOrFailure   = 447
	errorAllocationQuotaReached       = 486
	errorRoleConflict                 = 487
	errorServerError                  = 500
	errorInsufficientCapacity         = 508
)
const (
	attributeFamilyIPv4 = 0x01
	attributeFamilyIPV6 = 0x02
)

const (
	AttributeMappedAddress          = 0x0001
	AttributeResponseAddress        = 0x0002
	AttributeChangeRequest          = 0x0003
	AttributeSourceAddress          = 0x0004
	AttributeChangedAddress         = 0x0005
	AttributeUsername               = 0x0006
	AttributePassword               = 0x0007
	AttributeMessageIntegrity       = 0x0008
	AttributeErrorCode              = 0x0009
	AttributeUnknownAttributes      = 0x000a
	AttributeReflectedFrom          = 0x000b
	AttributeChannelNumber          = 0x000c
	AttributeLifetime               = 0x000d
	AttributeBandwidth              = 0x0010
	AttributeXorPeerAddress         = 0x0012
	AttributeData                   = 0x0013
	AttributeRealm                  = 0x0014
	AttributeNonce                  = 0x0015
	AttributeXorRelayedAddress      = 0x0016
	AttributeRequestedAddressFamily = 0x0017
	AttributeEvenPort               = 0x0018
	AttributeRequestedTransport     = 0x0019
	AttributeDontFragment           = 0x001a
	AttributeXorMappedAddress       = 0x0020
	AttributeTimerVal               = 0x0021
	AttributeReservationToken       = 0x0022
	AttributePriority               = 0x0024
	AttributeUseCandidate           = 0x0025
	AttributePadding                = 0x0026
	AttributeResponsePort           = 0x0027
	AttributeConnectionID           = 0x002a
	AttributeXorMappedAddressExp    = 0x8020
	AttributeSoftware               = 0x8022
	AttributeAlternateServer        = 0x8023
	AttributeCacheTimeout           = 0x8027
	AttributeFingerprint            = 0x8028
	AttributeIceControlled          = 0x8029
	AttributeIceControlling         = 0x802a
	AttributeResponseOrigin         = 0x802b
	AttributeOtherAddress           = 0x802c
	AttributeEcnCheckStun           = 0x802d
	AttributeCiscoFlowdata          = 0xc000
)

const (
	TypeBindingRequest                 = 0x0001
	TypeBindingResponse                = 0x0101
	TypeBindingErrorResponse           = 0x0111
	TypeSharedSecretRequest            = 0x0002
	TypeSharedSecretResponse           = 0x0102
	TypeSharedErrorResponse            = 0x0112
	TypeAllocate                       = 0x0003
	TypeAllocateResponse               = 0x0103
	TypeAllocateErrorResponse          = 0x0113
	TypeRefresh                        = 0x0004
	TypeRefreshResponse                = 0x0104
	TypeRefreshErrorResponse           = 0x0114
	TypeSend                           = 0x0006
	TypeSendResponse                   = 0x0106
	TypeSendErrorResponse              = 0x0116
	TypeData                           = 0x0007
	TypeDataResponse                   = 0x0107
	TypeDataErrorResponse              = 0x0117
	TypeCreatePermisiion               = 0x0008
	TypeCreatePermisiionResponse       = 0x0108
	TypeCreatePermisiionErrorResponse  = 0x0118
	TypeChannelBinding                 = 0x0009
	TypeChannelBindingResponse         = 0x0109
	TypeChannelBindingErrorResponse    = 0x0119
	TypeConnect                        = 0x000a
	TypeConnectResponse                = 0x010a
	TypeConnectErrorResponse           = 0x011a
	TypeConnectionBind                 = 0x000b
	TypeConnectionBindResponse         = 0x010b
	TypeConnectionBindErrorResponse    = 0x011b
	TypeConnectionAttempt              = 0x000c
	TypeConnectionAttemptResponse      = 0x010c
	TypeConnectionAttemptErrorResponse = 0x011c
)
