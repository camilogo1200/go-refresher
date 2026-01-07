# Description

- This folder exemplifies the use of feature flags to select golang source files based on feature flags using the
  directive //go:build <feature_flag>
- To avoid conflicts the files that shares the function signatures shpould have the inverse of the feature flag on the
  top level. I.E. //go:build !<feature_flag>
- The directive //go:build <feature_flag> should be always on top before the package declaration.
- The go command line assumes that when the feature flag is present the feature flag = true

## Note: 
- This is a TODO mini app to show the behavior using feature flags for build / compilation 



````bash
# this command includes the files that are marked with the //go:build feature_newui on the top of the file 
# feature_newui = true
go build -tags=feature_newui -o myapp ./cmd/api


````