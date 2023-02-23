<h1 align="center">
  <p align="center">Koordetector</p>
</h1>

English | [简体中文](./README-zh_CN.md)

## Introduction

Koordetector is a supportive project for Interference Detection feature of Koodinator. It aims to expand Koordinator's ability to collect performance metrics in the form of plug-ins, take part of responsibilities for performance metrics analysis, feeding back conclusions about whether some pods are interfered, and also play a sort of SIG-like role for exploring advanced metrics collection methods such as eBPF. 

Koordetector enhances the interference detection feature by dividing performance metrics into two types. One is metrics suitable for stand-alone collection and analysis on single node such as PSI, which is a percentage with a common abnormal threshold. The other is metrics that needs to be analyzed according to the workload itself such as CPI, which different applications have different value ranges. Koordetector handles the second type of metrics by a component named `Interference Manager`, which runs in the control plane, gathers and aggregates metrics by workloads, and analyzes them altogether using various strategies. Notice that the analysis conclusion of both types of metrics are feedback to Koordinator for further usage.    

Koordetector implements the above features by providing the following:

- Well-designed plug-in framework to manage metrics collecton tools from Koordinator, Koordetector and third-party.
- Complete and efficient data aggregation link to achieve both high analysis precision and acceptable overhead, with the help of histogram algorithms, sliding window, TSDB, Prometheus or custom metrics server, etc. 
- Intelligent interference detection algorithms and strategies, including simple empirical threshold method, ML, DL and so on. 
- A set of metrics collection tools and matching solution demos with documents, e.g., CPU schedule latency collector by eBPF with compatibility solution on different kernel versions.

![koordetector](docs/images/koordetector.svg)

## Code of conduct

The Koordetector is part of Koordinator community and is guided by our [Code of Conduct](CODE_OF_CONDUCT.md), which we encourage everybody to read
before participating.

In the interest of fostering an open and welcoming environment, we as contributors and maintainers pledge to making
participation in our project and our community a harassment-free experience for everyone, regardless of age, body size,
disability, ethnicity, level of experience, education, socio-economic status,
nationality, personal appearance, race, religion, or sexual identity and orientation.

## Contributing

You are warmly welcome to hack on Koordetector. We have prepared a detailed guide [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Koordetector is licensed under the Apache License, Version 2.0. See [LICENSE](./LICENSE) for the full license text.




