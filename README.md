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
Application draws two types of configuration files. General application config file "app.yaml":
```sh
sites: 2
stages: 1
uiengine: tview
```
And device and sequence specific config files placed in "config" directory. I.e:
```sh
hardware:
- site: 0
  device_name: testdevice
- site: 1
  device_name: testdevice
sequence:
- step_label: Send test command
  retry: 3
  device: testdevice
  timeout: 1000
  stepsettings:
      function: TestAction1
- step_label: Send test command2
  retry: 3
  device: testdevice
  timeout: 1000
  stepsettings:
      function: TestAction2
```
General configuration file determines number of sites that will be performing sequence of tasks and UI that will be loaded. Right now there is only termnial UI option written in tview. There is also option of running it without UI: in this case app will load "config.yaml" from "config" directory.
Specific config determines which modules will be loaded and which sites this module will work on.

To configure a device we would do something like this:
```sh
- site: 0
  device_name: genericuart
  settings:
    address: /dev/ttyUSB0
    baudrate: 9600
```
This piece configures *genericuart* device on site *0* with specified *settings*. As far as main application compnents are concerned, site and device name are the only things needed. Settings are used by a module itself and are parsed by *ConfigResolver.go*. In this file we determine what happens with *settings* section from the config.

Each sequence steps is configured like this:
```sh
- step_label: Send test command
  retry: 3
  device: testdevice
  timeout: 1000
  stepsettings:
      function: TestAction1
```
Main components care for sections: 
* *step_label* - Sets name of the step that is displayed in UI and log
* *retry* - Sets number of retries that application will perform if task results in fail
* *device* - Sets what module will receive event with this task in mind and should be the same as device name from hardware section
* *timeout* - Sets timeout constant - if module doesn't respond in that time, application resolves result as timeout error
* *step_settings* - sets things that are parsed nad resolved by module - it can contain function name and parameters that will be performed by module
<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- Data -->
## Reports and logs
All report data is stored locally in *reports.db* file in project directory. Application uses sqlite3 for this functionality. Log data is also stored in local db *log.db* created by sqlite3, it is also sent to UI component of the application.
<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- Prospect of developement -->
## Prospect of developement
Right now this project although functional has several aspects that draw it back and functionalities that could be implemented to improve it's stance as a proper base for functional testing platform:

Improvements to existing components:
* **Generating Reports** - Right now report generation is in it's early stages, it includes basic information and all sequence steps are stored as plaint text in one of the records. This can cause problems with longer sequences and makes it impossible to create report analyzing tools that could go into details on every step. Better approach would be to create One-Mant junction table with overall reports as one table and all steps in second table which would contain foreign key (report id).
* **Logging** - Logging is handled by main loop of the application and derives results passed by module to more readable form. Other approach would be to create *Logger* Singleton that could be instantiated troughout mainloop and modules alike making it more robust and detailed
* **Better error/exception handling** - Because in this kind of application continous operation is of essence errors that won't impact main event loop or UI operation won't result in panicking. This would call for custom error component that differentiates between different kind of errors for more consistent error handling

Functionality for addition:
* **Configuration** - Config files could be modified in UI component of the application. This would ensure that config files are properly formatted
* **More generic modules** - Although this project serves as a base, it could ship with more generic modules serving as base
* **Synchronization** - Most test sequencers ship with synchronization components that allow for better control over sequence execution between sites (i.e semaphores or sequence locks)
* **Stage separation** - Not all testing and sequence setups are symmetrical. This means that site0 and site1 can have different sequences because of different equipement (space constraints or functional separation). In this case we would be talking about "Stages" which are also a standard feature of test sequencers
<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->
## Contact

Piotr Snarski - snarski.piotrek@gmail.com

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->



