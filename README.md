# gzs

This is an attempt to extract my take on [soci](https://github.com/awslabs/soci-snapshotter) from my [dag.dev repo](https://github.com/jonjohnsonjr/dag.dev/tree/main/internal/soci).

This mostly exists to avoid shelling out or using cgo.
If you're looking for a tool that does this kind of thing, you are probably better off using [`gztool`](https://github.com/circulosmeos/gztool).

The interfaces and data formats in this are suboptimal and will change over time.

I've included a scrappy `gzs` tool under `cmd/gzs` that implements `index`, `ls`, and `cat` to verify that it kind of works.
