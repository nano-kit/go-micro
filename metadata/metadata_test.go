package metadata

import (
	"context"
	"reflect"
	"testing"
)

func TestMetadataSet(t *testing.T) {
	ctx := Set(context.TODO(), "Key", "val")

	val, ok := Get(ctx, "Key")
	if !ok {
		t.Fatal("key Key not found")
	}
	if val != "val" {
		t.Errorf("key Key with value val != %v", val)
	}
}

func TestMetadataDelete(t *testing.T) {
	md := Metadata{
		"Foo": "bar",
		"Baz": "empty",
	}

	ctx := NewContext(context.TODO(), md)
	ctx = Delete(ctx, "Baz")

	emd, ok := FromContext(ctx)
	if !ok {
		t.Fatal("key Key not found")
	}

	_, ok = emd["Baz"]
	if ok {
		t.Fatal("key Baz not deleted")
	}

}

func TestMetadataCopy(t *testing.T) {
	md := Metadata{
		"Foo": "bar",
		"bar": "baz",
	}

	cp := Copy(md)

	for k, v := range md {
		if cv := cp[k]; cv != v {
			t.Fatalf("Got %s:%s for %s:%s", k, cv, k, v)
		}
	}
}

func TestMetadataContext(t *testing.T) {
	md := Metadata{
		"Foo": "bar",
	}

	ctx := NewContext(context.TODO(), md)

	emd, ok := FromContext(ctx)
	if !ok {
		t.Errorf("Unexpected error retrieving metadata, got %t", ok)
	}

	if emd["Foo"] != md["Foo"] {
		t.Errorf("Expected key: %s val: %s, got key: %s val: %s", "Foo", md["Foo"], "Foo", emd["Foo"])
	}

	if i := len(emd); i != 1 {
		t.Errorf("Expected metadata length 1 got %d", i)
	}
}

func TestMergeContext(t *testing.T) {
	type args struct {
		existing  Metadata
		append    Metadata
		overwrite bool
	}
	tests := []struct {
		name string
		args args
		want Metadata
	}{
		{
			name: "matching key, overwrite false",
			args: args{
				existing:  Metadata{"Foo": "bar", "Sumo": "demo"},
				append:    Metadata{"Sumo": "demo2"},
				overwrite: false,
			},
			want: Metadata{"Foo": "bar", "Sumo": "demo"},
		},
		{
			name: "matching key, overwrite true",
			args: args{
				existing:  Metadata{"Foo": "bar", "Sumo": "demo"},
				append:    Metadata{"Sumo": "demo2"},
				overwrite: true,
			},
			want: Metadata{"Foo": "bar", "Sumo": "demo2"},
		},
		{
			name: "matching key, overwrite true",
			args: args{
				existing:  Metadata{"Foo": "bar", "Sumo": "demo"},
				append:    Metadata{"Sumo": "demo2", "Foo": "bar2"},
				overwrite: true,
			},
			want: Metadata{"Foo": "bar2", "Sumo": "demo2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := FromContext(MergeContext(NewContext(context.TODO(), tt.args.existing), tt.args.append, tt.args.overwrite)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeContext() = %v, want %v", got, tt.want)
			}
			ctx := NewContext(context.TODO(), Metadata{"Foo": "bar", "Sumo": "demo"})
			ctx = MergeContext(ctx, Metadata{"Sumo": "demo2", "Foo": "bar2"}, true)
			ctx = MergeContext(ctx, Metadata{"Sumo": "demo3", "Foo": "bar3"}, true)
			ctx = MergeContext(ctx, Metadata{"Sumo": "demo4", "Foo": "bar4"}, true)
			ctx = MergeContext(ctx, Metadata{"Sumo": "demo5", "Foo": "bar6"}, true)
			if got, _ := FromContext(ctx); !reflect.DeepEqual(got, Metadata{"Sumo": "demo5", "Foo": "bar6"}) {
				t.Errorf("MergeContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
