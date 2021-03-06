/*
Copyright 2019 The GitLab-Controller Authors.

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

package gitlab

import (
	"fmt"
	"testing"

	xpcorev1alpha1 "github.com/crossplaneio/crossplane/pkg/apis/core/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func Test_awsSecretConnectionCreator_create(t *testing.T) {
	type want struct {
		err  error
		data map[string][]byte
	}
	tests := map[string]struct {
		args map[string][]byte
		want want
	}{
		"Default": {
			args: nil,
			want: want{err: errors.New(errorMsgEmptyConnectionSecret)},
		},
		"Empty": {
			args: make(map[string][]byte),
			want: want{
				data: make(map[string][]byte),
				err:  errors.New(errorMsgEmptyConnectionSecret),
			},
		},
		"NoKeys": {
			args: map[string][]byte{"foo": []byte("bar")},
			want: want{
				data: map[string][]byte{
					"foo":         []byte("bar"),
					connectionKey: []byte(fmt.Sprintf(awsConnectionFmt, "", "")),
				},
			},
		},
		"Keys": {
			args: map[string][]byte{
				"foo": []byte("bar"),
				xpcorev1alpha1.ResourceCredentialsSecretUserKey:     []byte("test-access"),
				xpcorev1alpha1.ResourceCredentialsSecretPasswordKey: []byte("test-secret"),
			},
			want: want{
				data: map[string][]byte{
					"foo": []byte("bar"),
					xpcorev1alpha1.ResourceCredentialsSecretUserKey:     []byte("test-access"),
					xpcorev1alpha1.ResourceCredentialsSecretPasswordKey: []byte("test-secret"),
					connectionKey: []byte(fmt.Sprintf(awsConnectionFmt, "test-access", "test-secret")),
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			creator := &awsSecretConnectionCreator{}
			secret := &corev1.Secret{Data: tt.args}
			if diff := cmp.Diff(creator.create(secret), tt.want.err, cmpErrors); diff != "" {
				t.Errorf("awsSecretConnectionCreator.create() error = %s", diff)
				return
			}
			if diff := cmp.Diff(secret.Data, tt.want.data); tt.want.data != nil && diff != "" {
				t.Errorf("awsSecretConnectionCreator.create() data %s", diff)
			}
		})
	}
}

func Test_awsSecretS3CmdConfigCreator_create(t *testing.T) {
	type want struct {
		data map[string][]byte
		err  error
	}
	tests := map[string]struct {
		args map[string][]byte
		want want
	}{
		"Default": {
			args: nil,
			want: want{err: errors.New(errorMsgEmptyConnectionSecret)},
		},
		"Empty": {
			args: make(map[string][]byte),
			want: want{
				data: make(map[string][]byte),
				err:  errors.New(errorMsgEmptyConnectionSecret),
			},
		},
		"NoKeys": {
			args: map[string][]byte{"foo": []byte("bar")},
			want: want{
				data: map[string][]byte{
					"foo":     []byte("bar"),
					configKey: []byte(fmt.Sprintf(awsS3CmdConfigFmt, "", "")),
				},
			},
		},
		"Keys": {
			args: map[string][]byte{
				"foo": []byte("bar"),
				xpcorev1alpha1.ResourceCredentialsSecretUserKey:     []byte("test-access"),
				xpcorev1alpha1.ResourceCredentialsSecretPasswordKey: []byte("test-secret"),
			},
			want: want{
				data: map[string][]byte{
					"foo": []byte("bar"),
					xpcorev1alpha1.ResourceCredentialsSecretUserKey:     []byte("test-access"),
					xpcorev1alpha1.ResourceCredentialsSecretPasswordKey: []byte("test-secret"),
					configKey: []byte(fmt.Sprintf(awsS3CmdConfigFmt, "test-access", "test-secret")),
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			creator := &awsSecretS3CmdConfigCreator{}
			secret := &corev1.Secret{Data: tt.args}
			if diff := cmp.Diff(creator.create(secret), tt.want.err, cmpErrors); diff != "" {
				t.Errorf("awsSecretS3CmdConfigCreator.create() error %s", diff)
			}
			if diff := cmp.Diff(secret.Data, tt.want.data); diff != "" {
				t.Errorf("awsSecretS3CmdConfigCreator.create() data %s", diff)
			}
		})
	}

}

func Test_awsSecretUpdater_update(t *testing.T) {
	testError := errors.New("test-error")
	type fields struct {
		connection secretDataCreator
		config     secretDataCreator
	}
	tests := map[string]struct {
		fields  fields
		args    *corev1.Secret
		wantErr error
	}{
		"CreateConnectionFailed": {
			fields: fields{
				connection: &mockSecretDataCreator{
					mockCreate: func(secret *corev1.Secret) error {
						return testError
					},
				},
				config: &mockSecretDataCreator{
					mockCreate: func(secret *corev1.Secret) error {
						return nil
					},
				},
			},
			wantErr: errors.Wrapf(testError, errorFailedToCreateConnectionData),
		},
		"CreateConfigFailed": {
			fields: fields{
				connection: &mockSecretDataCreator{
					mockCreate: func(secret *corev1.Secret) error {
						return nil
					},
				},
				config: &mockSecretDataCreator{
					mockCreate: func(secret *corev1.Secret) error {
						return testError
					},
				},
			},
			wantErr: errors.Wrapf(testError, errorFailedToCreateConfigData),
		},
		"Successful": {
			fields: fields{
				connection: &mockSecretDataCreator{
					mockCreate: func(secret *corev1.Secret) error {
						return nil
					},
				},
				config: &mockSecretDataCreator{
					mockCreate: func(secret *corev1.Secret) error {
						return nil
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			u := &awsSecretUpdater{
				connection: tt.fields.connection,
				config:     tt.fields.config,
			}
			if diff := cmp.Diff(u.update(tt.args), tt.wantErr, cmpErrors); diff != "" {
				t.Errorf("awsSecretUpdater.update() error %s", diff)
			}
		})
	}
}
