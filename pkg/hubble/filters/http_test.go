// Copyright 2019-2020 Authors of Hubble
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !privileged_tests

package filters

import (
	"context"
	"testing"

	pb "github.com/cilium/cilium/api/v1/flow"
	v1 "github.com/cilium/cilium/pkg/hubble/api/v1"
	"github.com/cilium/cilium/pkg/monitor/api"
)

func TestHttpStatusCodeFilter(t *testing.T) {
	httpFlow := func(http *pb.HTTP) *v1.Event {
		return &v1.Event{
			Event: &pb.Flow{
				EventType: &pb.CiliumEventType{
					Type: api.MessageTypeAccessLog,
				},
				L7: &pb.Layer7{
					Record: &pb.Layer7_Http{
						Http: http,
					},
				}},
		}
	}

	type args struct {
		f  []*pb.FlowFilter
		ev []*v1.Event
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []bool
	}{
		{
			name: "status code full",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{"200", "302"},
						EventType:      []*pb.EventTypeFilter{{Type: api.MessageTypeAccessLog}},
					},
				},
				ev: []*v1.Event{
					httpFlow(&pb.HTTP{Code: 200}),
					httpFlow(&pb.HTTP{Code: 302}),
					httpFlow(&pb.HTTP{Code: 404}),
					httpFlow(&pb.HTTP{Code: 500}),
				},
			},
			want: []bool{
				true,
				true,
				false,
				false,
			},
			wantErr: false,
		},
		{
			name: "status code prefix",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{"40+", "5+"},
						EventType:      []*pb.EventTypeFilter{{Type: api.MessageTypeAccessLog}},
					},
				},
				ev: []*v1.Event{
					httpFlow(&pb.HTTP{Code: 302}),
					httpFlow(&pb.HTTP{Code: 400}),
					httpFlow(&pb.HTTP{Code: 404}),
					httpFlow(&pb.HTTP{Code: 410}),
					httpFlow(&pb.HTTP{Code: 004}),
					httpFlow(&pb.HTTP{Code: 500}),
					httpFlow(&pb.HTTP{Code: 501}),
					httpFlow(&pb.HTTP{Code: 510}),
					httpFlow(&pb.HTTP{Code: 050}),
				},
			},
			want: []bool{
				false,
				true,
				true,
				false,
				false,
				true,
				true,
				true,
				false,
			},
			wantErr: false,
		},
		{
			name: "invalid data",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{"200"},
						EventType:      []*pb.EventTypeFilter{{Type: api.MessageTypeAccessLog}},
					},
				},
				ev: []*v1.Event{
					{Event: &pb.Flow{}},
					httpFlow(&pb.HTTP{}),
					httpFlow(&pb.HTTP{Code: 777}),
				},
			},
			want: []bool{
				false,
				false,
				false,
			},
			wantErr: false,
		},
		{
			name: "invalid empty filter",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{""},
						EventType:      []*pb.EventTypeFilter{{Type: api.MessageTypeAccessLog}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid catch-all prefix",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{"+"},
						EventType:      []*pb.EventTypeFilter{{Type: api.MessageTypeAccessLog}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status code",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{"909"},
						EventType:      []*pb.EventTypeFilter{{Type: api.MessageTypeAccessLog}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status code prefix",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{"3++"},
						EventType:      []*pb.EventTypeFilter{{Type: api.MessageTypeAccessLog}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status code prefix",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{"3+0"},
						EventType:      []*pb.EventTypeFilter{{Type: api.MessageTypeAccessLog}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty event type filter",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{"200"},
						EventType:      []*pb.EventTypeFilter{},
					},
				},
				ev: []*v1.Event{
					httpFlow(&pb.HTTP{Code: 200}),
				},
			},
			want: []bool{
				true,
			},
			wantErr: false,
		},
		{
			name: "compatible event type filter",
			args: args{
				f: []*pb.FlowFilter{
					{
						HttpStatusCode: []string{"200"},
						EventType: []*pb.EventTypeFilter{
							{Type: api.MessageTypeAccessLog},
							{Type: api.MessageTypeTrace},
						},
					},
				},
				ev: []*v1.Event{
					httpFlow(&pb.HTTP{Code: 200}),
				},
			},
			want: []bool{
				true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl, err := BuildFilterList(context.Background(), tt.args.f, []OnBuildFilter{&HTTPFilter{}})
			if (err != nil) != tt.wantErr {
				t.Errorf("\"%s\" error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			for i, ev := range tt.args.ev {
				if got := fl.MatchOne(ev); got != tt.want[i] {
					t.Errorf("\"%s\" got %d = %v, want %v", tt.name, i, got, tt.want[i])
				}
			}
		})
	}
}
