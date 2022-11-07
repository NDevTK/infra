package spec

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"
)

func TestCreateParser(t *testing.T) {
	Convey("Singe Create", t, func() {
		p, err := newCreateParser("linux-amd64", []*Spec_Create{
			{
				Source: &Spec_Create_Source{
					Method: &Spec_Create_Source_Url{
						Url: &UrlSource{
							DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
							Version:     "1.2.12",
						},
					},
					UnpackArchive:  true,
					CpeBaseAddress: "cpe:/a:zlib:zlib",
				},
				Build: &Spec_Create_Build{},
			},
		})
		So(err, ShouldBeNil)
		So(proto.Equal(p.create, &Spec_Create{
			Source: &Spec_Create_Source{
				Method: &Spec_Create_Source_Url{
					Url: &UrlSource{
						DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
						Version:     "1.2.12",
					},
				},
				UnpackArchive:  true,
				CpeBaseAddress: "cpe:/a:zlib:zlib",
			},
			Build: &Spec_Create_Build{},
		}), ShouldBeTrue)
	})

	Convey("Multiple Create", t, func() {
		p, err := newCreateParser("linux-amd64", []*Spec_Create{
			{
				Source: &Spec_Create_Source{
					Method: &Spec_Create_Source_Url{
						Url: &UrlSource{
							DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
						},
					},
					UnpackArchive: true,
				},
				Build: &Spec_Create_Build{},
			},
			{
				Source: &Spec_Create_Source{
					Method: &Spec_Create_Source_Url{
						Url: &UrlSource{
							Version: "1.2.12",
						},
					},
					CpeBaseAddress: "cpe:/a:zlib:zlib",
				},
			},
		})
		So(err, ShouldBeNil)
		So(proto.Equal(p.create, &Spec_Create{
			Source: &Spec_Create_Source{
				Method: &Spec_Create_Source_Url{
					Url: &UrlSource{
						DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
						Version:     "1.2.12",
					},
				},
				UnpackArchive:  true,
				CpeBaseAddress: "cpe:/a:zlib:zlib",
			},
			Build: &Spec_Create_Build{},
		}), ShouldBeTrue)
	})

	Convey("Match Platform", t, func() {
		p, err := newCreateParser("linux-amd64", []*Spec_Create{
			{
				PlatformRe: "linux-.*",
				Source: &Spec_Create_Source{
					Method: &Spec_Create_Source_Url{
						Url: &UrlSource{
							DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
							Version:     "1.2.12",
						},
					},
					UnpackArchive:  true,
					CpeBaseAddress: "cpe:/a:zlib:zlib",
				},
				Build: &Spec_Create_Build{},
			},
			{
				PlatformRe:  "unknown-.*",
				Unsupported: true,
			},
		})
		So(err, ShouldBeNil)
		So(proto.Equal(p.create, &Spec_Create{
			Source: &Spec_Create_Source{
				Method: &Spec_Create_Source_Url{
					Url: &UrlSource{
						DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
						Version:     "1.2.12",
					},
				},
				UnpackArchive:  true,
				CpeBaseAddress: "cpe:/a:zlib:zlib",
			},
			Build: &Spec_Create_Build{},
		}), ShouldBeTrue)
	})

	Convey("Unsupported Platform Explicit", t, func() {
		_, err := newCreateParser("linux-amd64", []*Spec_Create{
			{
				Unsupported: true,
			},
			{
				Source: &Spec_Create_Source{
					Method: &Spec_Create_Source_Url{
						Url: &UrlSource{
							DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
							Version:     "1.2.12",
						},
					},
					UnpackArchive:  true,
					CpeBaseAddress: "cpe:/a:zlib:zlib",
				},
				Build: &Spec_Create_Build{},
			},
		})
		So(err, ShouldEqual, ErrPackageNotAvailable)
	})

	Convey("Unsupported Platform Implicit", t, func() {
		_, err := newCreateParser("linux-amd64", []*Spec_Create{
			{
				PlatformRe: "unknown-.*",
				Source: &Spec_Create_Source{
					Method: &Spec_Create_Source_Url{
						Url: &UrlSource{
							DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
							Version:     "1.2.12",
						},
					},
					UnpackArchive:  true,
					CpeBaseAddress: "cpe:/a:zlib:zlib",
				},
				Build: &Spec_Create_Build{},
			},
		})
		So(err, ShouldEqual, ErrPackageNotAvailable)
	})

	Convey("Merge Values", t, func() {
		p, err := newCreateParser("linux-amd64", []*Spec_Create{
			{
				Source: &Spec_Create_Source{
					Method: &Spec_Create_Source_Url{
						Url: &UrlSource{
							DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
							Version:     "1.2.12",
						},
					},
					UnpackArchive:  true,
					PatchDir:       []string{"patches1", "patches2"},
					CpeBaseAddress: "cpe:/a:zlib:zlib",
				},
				Build: &Spec_Create_Build{},
			},
			{
				Source: &Spec_Create_Source{
					PatchDir:       []string{"patches1"},
					CpeBaseAddress: "cpe:/a:zlib:zlib1",
				},
			},
		})
		So(err, ShouldBeNil)
		So(proto.Equal(p.create, &Spec_Create{
			Source: &Spec_Create_Source{
				Method: &Spec_Create_Source_Url{
					Url: &UrlSource{
						DownloadUrl: "https://zlib.net/fossils/zlib-1.2.12.tar.gz",
						Version:     "1.2.12",
					},
				},
				UnpackArchive:  true,
				PatchDir:       []string{"patches1"},
				CpeBaseAddress: "cpe:/a:zlib:zlib1",
			},
			Build: &Spec_Create_Build{},
		}), ShouldBeTrue)
	})
}
