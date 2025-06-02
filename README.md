<a name="readme-top"></a>


<!-- ABOUT THE PROJECT -->
## About The Project

Checkerbox is lightweight and minimal hardware testing and measurement application simillar to <a href="https://www.ni.com/en/shop/electronic-test-instrumentation/application-software-for-electronic-test-and-instrumentation-category/what-is-teststand.html">NI TestStand</a> serving as base for further expansion. Application works on hybrid concurrent model. Devices that are used to perform tests and/or measurements are 
written as modules that have their own event-loop waiting for events sent by main-loop which will determine what actions this device will take. Event data structure provides address for channel where result will be sent - this way situation where devices have to be provided with returning event-bus and care to not access it while other module is in control can be avoided. What devices will be initialized and what are test sequence steps is determined
by configuration files written in <a href="https://yaml.org/">YAML</a> data format with clear and easy to understand way.
<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

* <a href="https://go.dev/">GO</a>
* <a href="https://github.com/rivo/tview">Tview</a>
* <a href="https://www.sqlite.org/">SQLite</a>
* <a href="https://gorm.io/index.html">GORM</a>



<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- DESIGN PRINCIPLES -->
## Design principles
Project serves as basis for further developement and expansion. Giving lightweight and fast core built with concurrency in mind.

<!-- USAGE EXAMPLES -->
## Usage
Using this software is as simple as it gets. Golang in version 1.23 or higher is required.
Clone this repository with:
```sh
git clone https://github.com/H33Kz/checkerbox
```
Into your local directory. Run software inside main project directory by:
```sh
go run .
```
Or
```sh
go build .
```
Note that for some functionality like accessing serial port address (Which is required by one of the example modules) needs running this application as and administrator.
<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- Config -->
## Configuration

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- Data -->
## Reports and logs

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- Prospect of developement -->
## Prospect of developement
<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->
## Contact

Piotr Snarski - snarski.piotrek@gmail.com

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->



