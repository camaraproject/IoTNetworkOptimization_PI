/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// MarshalBSON implements bson.Marshaler for Device.
func (d Device) MarshalBSON() ([]byte, error) {
	doc := bson.M{}

	if d.Ipv4Address != nil {
		doc["ipv4Address"] = d.Ipv4Address
	}
	if d.Ipv6Address != nil {
		doc["ipv6Address"] = *d.Ipv6Address
	}
	if d.NetworkAccessIdentifier != nil {
		doc["networkAccessIdentifier"] = *d.NetworkAccessIdentifier
	}
	if d.PhoneNumber != nil {
		doc["phoneNumber"] = *d.PhoneNumber
	}

	return bson.Marshal(doc)
}

// UnmarshalBSON implements bson.Unmarshaler for Device.
func (d *Device) UnmarshalBSON(data []byte) error {
	doc := bson.M{}
	if err := bson.Unmarshal(data, &doc); err != nil {
		return err
	}

	if v, ok := doc["ipv4Address"]; ok {
		bytes, err := bson.Marshal(v)
		if err != nil {
			return err
		}
		var ipv4 DeviceIpv4Addr
		if err := bson.Unmarshal(bytes, &ipv4); err != nil {
			return err
		}
		d.Ipv4Address = &ipv4
	}

	if v, ok := doc["ipv6Address"]; ok {
		if str, ok := v.(string); ok {
			ipv6 := DeviceIpv6Address(str)
			d.Ipv6Address = &ipv6
		}
	}

	if v, ok := doc["networkAccessIdentifier"]; ok {
		if str, ok := v.(string); ok {
			nai := NetworkAccessIdentifier(str)
			d.NetworkAccessIdentifier = &nai
		}
	}

	if v, ok := doc["phoneNumber"]; ok {
		if str, ok := v.(string); ok {
			phone := PhoneNumber(str)
			d.PhoneNumber = &phone
		}
	}

	return nil
}

// MarshalBSON implements bson.Marshaler for DeviceIpv4Addr.
func (d DeviceIpv4Addr) MarshalBSON() ([]byte, error) {
	doc := bson.M{}
	if d.PrivateAddress != nil {
		doc["privateAddress"] = *d.PrivateAddress
	}
	if d.PublicAddress != nil {
		doc["publicAddress"] = *d.PublicAddress
	}
	if d.PublicPort != nil {
		doc["publicPort"] = *d.PublicPort
	}
	return bson.Marshal(doc)
}

// UnmarshalBSON implements bson.Unmarshaler for DeviceIpv4Addr.
func (d *DeviceIpv4Addr) UnmarshalBSON(data []byte) error {
	doc := bson.M{}
	if err := bson.Unmarshal(data, &doc); err != nil {
		return err
	}

	if v, ok := doc["privateAddress"]; ok {
		if addr, ok := v.(string); ok {
			singleAddr := SingleIpv4Addr(addr)
			d.PrivateAddress = &singleAddr
		}
	}

	if v, ok := doc["publicAddress"]; ok {
		if addr, ok := v.(string); ok {
			singleAddr := SingleIpv4Addr(addr)
			d.PublicAddress = &singleAddr
		}
	}

	if v, ok := doc["publicPort"]; ok {
		if port, ok := v.(int32); ok {
			portInt := Port(port)
			d.PublicPort = &portInt
		} else if port, ok := v.(int); ok {
			portInt := Port(port)
			d.PublicPort = &portInt
		}
	}

	return nil
}
