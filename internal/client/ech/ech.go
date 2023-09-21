// Package ech is responsible for implementing the Encrypted ClientHello logic.
package ech

import (
	"crypto/tls"
	"net"

	ctls "github.com/ameshkov/cfcrypto/tls"
	"github.com/ameshkov/gocurl/internal/output"
)

// HandshakeECH attempts to establish a ECH-enabled connection using the
// specified echConfigs.
//
// A few things about tlsConfig that is passed to it:
// ServerName will be used in the inner ClientHello.  For the outer ClientHello
// it will attempt to use the "public name" field of the ECH configuration.
// Regarding the multiple ECHConfig passed, it chooses the first with a suitable
// cipher suite which effectively means that it will almost always simply use
// the first ECHConfig from the slice.
func HandshakeECH(
	conn net.Conn,
	echConfigs []ctls.ECHConfig,
	tlsConfig *tls.Config,
	out *output.Output,
) (tlsConn net.Conn, err error) {
	out.Debug("Attempting to establish a ECH-enabled connection")

	// Copying the original tls config fields to ECH-enabled one.
	conf := &ctls.Config{
		ServerName:         tlsConfig.ServerName,
		MinVersion:         tlsConfig.MinVersion,
		MaxVersion:         tlsConfig.MaxVersion,
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
		NextProtos:         tlsConfig.NextProtos,
		ECHEnabled:         true,
		ClientECHConfigs:   echConfigs,
	}

	c := ctls.Client(conn, conf)
	err = c.Handshake()

	if err != nil {
		return nil, err
	}

	out.Debug("ECH-enabled connection has been established successfully")

	return c, nil
}
