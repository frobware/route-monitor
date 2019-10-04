# Monitor routes for reachability

## build

    $ make
	
## run

```console
$ ./bin/route-monitor                                                                                                                                                                                                
I1004 18:28:21.465277   29581 main.go:100] starting route controller                                                                                                                                                 
I1004 18:28:21.965756   29581 main.go:109] existing route: namespace: "openshift-authentication", host: "oauth-openshift.apps.amcdermo.devcluster.openshift.com"                                                     
I1004 18:28:21.965811   29581 main.go:109] existing route: namespace: "openshift-console", host: "downloads-openshift-console.apps.amcdermo.devcluster.openshift.com"                                                
I1004 18:28:21.965835   29581 main.go:109] existing route: namespace: "openshift-monitoring", host: "prometheus-k8s-openshift-monitoring.apps.amcdermo.devcluster.openshift.com"                                     
E1004 18:28:21.965861   29581 main.go:118] route "foo.com" does not exist                                                                                                                                            
I1004 18:28:22.216915   29581 main.go:134] route "https://downloads-openshift-console.apps.amcdermo.devcluster.openshift.com" IS reachable                                                                           
I1004 18:28:23.217237   29581 main.go:109] existing route: namespace: "openshift-monitoring", host: "grafana-openshift-monitoring.apps.amcdermo.devcluster.openshift.com"                                            
I1004 18:28:23.217291   29581 main.go:109] existing route: namespace: "openshift-authentication", host: "oauth-openshift.apps.amcdermo.devcluster.openshift.com"                                                     
I1004 18:28:23.217314   29581 main.go:109] existing route: namespace: "openshift-console", host: "downloads-openshift-console.apps.amcdermo.devcluster.openshift.com"                                                
E1004 18:28:23.217337   29581 main.go:118] route "foo.com" does not exist                                                                                                                                            
I1004 18:28:23.394278   29581 main.go:134] route "https://downloads-openshift-console.apps.amcdermo.devcluster.openshift.com" IS reachable                                                                           
I1004 18:28:24.394622   29581 main.go:109] existing route: namespace: "openshift-authentication", host: "oauth-openshift.apps.amcdermo.devcluster.openshift.com"                                                     
I1004 18:28:24.395471   29581 main.go:109] existing route: namespace: "openshift-console", host: "downloads-openshift-console.apps.amcdermo.devcluster.openshift.com"                                                
I1004 18:28:24.395505   29581 main.go:109] existing route: namespace: "openshift-monitoring", host: "prometheus-k8s-openshift-monitoring.apps.amcdermo.devcluster.openshift.com"                                     
E1004 18:28:24.395544   29581 main.go:118] route "foo.com" does not exist    
```
