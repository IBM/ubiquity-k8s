# Bluemix CLI Plugin Developer's Guide

This guide introduces how to develop a Bluemix CLI plugin by using utilities and libraries provided by Bluemix CLI SDK. It also covers specifications including wording, format and color of the terminal output that we highly recommend developers to follow.

## 1. Plugin Context Management

Bluemix CLI SDK provides a set of APIs to register and manage plugins. It is similar to  CF plugin interface, but has more Bluemix specific extensions.

### 1.1. Register a New Plugin

_Step 1:_ Define a new struct for Bluemix CLI plugin and associate Run and GetMetadata methods to that struct:

```go
import (
    "github.ibm.com/Bluemix/bluemix-cli-sdk/plugin"
)

type DemoPlugin struct {}

func (demo *DemoPlugin) Run(context plugin.PluginContext, args []string) {}

func (demo *DemoPlugin) GetMetadata() plugin.PluginMetadata {}
```

_Step 2:_ In main function, invoke plugin.Start method to register the plugin:

```go
func main() {
    plugin.Start(new(DemoPlugin))
}
```

_Step 3:_ Return a plugin.PluginMetadata struct to finalize the registration of the plugin in GetMetadata method:

```go
func (demo *DemoPlugin) GetMetadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name: "demo-plugin",
        Version: plugin.VersionType{
            Major: 1,
            Minor: 0,
            Build: 0,
        },
        MinCliVersion: plugin.VersionType{
            Major: 0,
            Minor: 0,
            Build: 1,
        },
        Commands: []plugin.Command{
            {
                Name:        "echo",
                Alias:       "ec",
                Description: "Echo a message on terminal.",
                Usage:       "bluemix echo MESSAGE [-u]",
                Flags: []plugin.Flag{
                    {
                        Name: "u",
                        Description: "Change the message to upper case.",
                        HasValue: false,
                    },
                },
            },
        },
    }
}
```

Below is the explanation of some of the fields in plugin.PluginMetadata struct:

- Name: The name of plugin. It will be displayed when using `bluemix plugin list` command or can be used to uninstall the plugin through `bluemix plugin uninstall` command.
- Version: The version of plugin.
- MinCliVersion: The minimal version of Bluemix CLI required by the plugin.
- Commands: The array of plugin.Commands to register the plugin commands.
- Alias: Alias of the Alias usually is a short name of the command.
- Command.Flags: The command flags (options) which will be displayed as a part of help output of the command.

_Step 4:_ Add the logic of plugin command process in Run method, for example:

```go
func (demo *DemoPlugin) Run(context plugin.PluginContext, args []string) {
    if args[0] == "echo" {
        // echo command logic here
    }
}
```

PluginContext provides the most useful methods which allow you to get command line properties from CF configuration as well as Bluemix specific properties.

### 1.2. Namespace

Bluemix CLI introduced a new concept called "Namespace". A namespace is a category of commands which have similar functionality. Some namespaces are predefined by Bluemix CLI and can be shared by plugins, others are non-shared namespaces which can be defined in each plugin. Plugin can reference a predefined namespace in Bluemix CLI or define a non-shared namespace by its own.

The following are shared-namespaces currently predefined in Bluemix CLI:

| Namespace | Description |
| --- | --- |
| iam | Manage accounts, orgs, spaces and users |
| catalog | Manage Bluemix catalog |
| app | Manage Bluemix applications |
| service | Manage Bluemix services |
| network | Manage network settings including region, domain and route|
| security | Manage security settings |
| bss | Retrieve usage and billing information |
| cf | Run Cloud Foundry CLI with Bluemix context |
| plugin | Manage plugin repositories and plugins |

For shared namespace, it can be referenced by plugin as the following example:

```go
import "github.ibm.com/Bluemix/bluemix-cli-sdk/plugin"

func (p *CatalogExtPlugin) GetMetadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        ...
        Commands: []plugin.Command{
            {
                Namespace:   "catalog",
                Name:        "cmd1",
            },
            {
                Namespace:   "catalog",
                Name:        "cmd2",
            },
        },
    }
}

func (p *CatalogExtPlugin) Run(context plugin.PluginContext, args []string) {
    switch args[0] {
    case "cmd1":
        // cmd1 command logic here
    case "cmd2":
        // cmd2 command logic here
    default:
        // command not recognized
    }
}
```

For non-shared namespace, it can be defined in plugin like:

```go
import "github.ibm.com/Bluemix/bluemix-cli-sdk/plugin"


func (p *DemoPlugin) GetMetadata() plugin.PluginMetadata {
    // define plugin namespace
    demo := plugin.Namespace{
        Name: "demo",
        Description: "For demonstration purpose",
    }

    return plugin.PluginMetadata{
        ...
        Namespaces: []plugin.Namespace {
            demo,
        },
        Commands: []plugin.Command{
            {
                Namespace:   demo.Name,
                Name:        "start",
                
            },
            {
                Namespace:   demo.Name,
                Name:        "help",
                
            },
            {
                Namespace:   demo.Name,
                Name:        "*",
            },
        },
    }
}

func (p *DemoPlugin) Run(context plugin.PluginContext, args []string) {
    switch args[0] {
    case "start":
        // start command logic here
    case "help"
        // print help
    default:
        // default run. Wildcard "*" matched
    }
}
```

The following items should be noticed here:

- The command registered in plugin can be associated with at most one namespace.
- If a command is not associated with any namespace, it will be registered as a root command. For example, if a command is associated with the network namespace, it must be executed by typing "bluemix network xxx", but if no namespace is associated with a command, the command will be invoked by "bluemix xxx" directly.
- All shared namespaces can only be defined in Bluemix CLI and referenced by plugins.
- Developer can define multiple non-shared namespaces in one plugin.
- If the plugin only belongs to non-shared namespaces, a special command with name "\*" can be defined, which means even if a user typed in a command which is not defined in current namespace, the plugin will still be invoked, otherwise, an error message will be displayed by Bluemix CLI. Taking the above code snippet as an example, if user typed in "bluemix demo others", then "default" logic in "Run" method will be hit.
- If multiple plugins define a namespace with the same name, they will follow "first install first serve strategy", which means the latter plugins can't be installed successfully due to the namespace conflict.
- Developer can define help command for non-shared namespace. If it is not defined, the help content will be generated by Bluemix CLI automatically based on registered commands. For shared namespaces the help command was predefined in Bluemix CLI.
- **Important:** No matter whether the command belongs to a namespace, args[0] will always be the command name or alias instead of the namespace name. You can get the namespace via `PluginContext.CommandNamespace()`.

### 1.3. Manage Plugin's Configuration

Bluemix CLI SDK provides APIs to allow you to access the plugin's own configuration saved in JSON format. Each plugin will have it's own configuration file be created automatically on installation. Take the following code as an example:

```go
func (demo *DemoPlugin) Run(context plugin.PluginContext, args []string){
    config := context.PluginConfig()

    // set
    err := config.Set("s", "string value")
    panic(err)

    err = config.Set("n", 123)
    panic(err)

    err = config.Set("ss", []string("foo", "bar"))
    panic(err)

    err = config.Set("m", nap[string]string{"one": "foo", "two": "bar"}})
    panic(err)

    // get
    myStr, err :=  config.GetString("s")
    panic(err)

    myNum, err := config.GetIntWithDefault("n", 100)
    panic(err)

    mySlice, err := config.GetStringSlice("ss")
    panic(err)

    myMap, err := config.GetStringMapString("m")
    panic(err)
    ...
}
```

# 2. Wording, Format and Color of Output

To keep user experience consistent, developers of Bluemix CLI plugin should apply specific wordings, formats and colors to the terminal output. Bluemix CLI SDK provides the utility to help plugin developers easily format and colorize the message output. We strongly recommend developers to comply with the following specifications so that the plugins are consistent with each other in terms of user experience. 

We tried to make the formats and colors to be similar to CF CLI too, to make the experience to be similar to CF as much as possible. 

### 2.1. Global Specification

1. Do NOT use "Please" in any message.
  Correct:
  <pre class="bx-console-block">Use '<span class="yellow">bluemix login</span>' to login first.</pre>
  Invalid:
  <pre class="bx-console-block">Please use '<span class="yellow">bluemix login</span>' to login first.</pre>
2. Capitalize the first letter for all sentences and short descriptions. For example:
  <pre class="bx-console-block">Change the instance count for an app or container group.</pre>
  <pre class="bx-console-block">-i   Number of instances</pre>
3. Add "..." at the end of "in-progress" messages. For example:
  <pre class="bx-console-block">Scaling container group '<span class="cyan">xxx</span>'...</pre>
4. Use "plug-in" instead of "plugin" in all places.

### 2.2. Plugin and Command Name

To name the plugin for a service

- Use full name of the service in all lower case, and hyphen to replace space, for example 'my-service'
- Use abbreviation only when it’s commonly accepted, for example use 'vpn' for Virtual Private Network service

To name the commands:

- Use lower case words and hyphen
- Follow a “noun-verb” sequence
- For command that list objects, use plural of the object name (or "object-list” if plural won’t work)
- Use common verbs like add, create, bind, update, delete …

For example, the following command names are correct:

route-map

route-unmap

template-run

plugin-repo-add

plugin-repo-remove

templates

...

The following command names are invalid:

Map-Route

map-route

map route

route map

route\_map

add-plugin-repo

plugin-add-repo

...

Use single quotation marks (') around the command in output message and the command itself should be yellow with **bold**. For example:

```
You have not logged in. Use 'bluemix login'to login first.
```


You can use the following APIs provided by Bluemix CLI SDK to print the above message:

```go
import "github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/terminal"

ui := terminal.NewStdUI()

ui.Say(`You have not logged in. Use "%s" to login first.`, 
    terminal.CommandColor("bluemix login"))

```

### 2.3. Entity Name

Add single quotation marks (') around entity names and keep the entity names in cyan with **bold**. For example:

```
Mapping route 'my-app.ng.bluemix.net' to CF application 'my-app'...
```

The Bluemix CLI SDK also provides API to help you print the above message:

```go
ui.Say("Mapping route '%s' to CF application '%s'...",
    terminal.EntityNameColor("my-app.ng.bluemix.net"),
    terminal.EntityNameColor("my-app"),
)
```

### 2.4. Help of Command

In the help content of plugin command, use upper case for all command variables and brackets "[]" around optional command options. For required options, do not wrap brackets around. The following is an example of the out put of the "help" command:

```
NAME:
   scale - Change the instance count for an app or container group
USAGE:
   bx scale RESOURCE_NAME [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]
   RESOURCE_NAME is the name of app or container group to be scaled
OPTIONS:
   -i value  Number of instances
   -k value  Disk limit (e.g. 256M, 1024M, 1G), valid only for scaling an app, not a container group
   -m value  Memory limit (e.g. 256M, 1024M, 1G), valid only for scaling an app, not a container group
   -f        Force restart of CF application without prompt, valid only for scaling an app, not a container group
```

### 2.5. Incorrect Usage

When user uses the command with wrong usage (e.g. incorrect number of arguments, invalid option value, required options not specified and etc.), the message in following format should be displayed:

```
Incorrect usage.
NAME:
   scale - Change the instance count for an app or container group
USAGE:
   bx scale RESOURCE_NAME [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]
   RESOURCE_NAME is the name of app or container group to be scaled
OPTIONS:
   -i value  Number of instances
   -k value  Disk limit (e.g. 256M, 1024M, 1G), valid only for scaling an app, not a container group
   -m value  Memory limit (e.g. 256M, 1024M, 1G), valid only for scaling an app, not a container group
   -f        Force restart of CF application without prompt, valid only for scaling an app, not a container group
```

### 2.6. Command Failure

If a command failed due to the client-side or server-side error, the following output format need to be complied with:

```
Creating application 'my-app'...
FAILED
An application with name 'my-app' already exists.
Use another name and try again.
```

```
Scaling container group 'xxx'...
FAILED
A server error occurred while scaling the container group.
Try again later. If the problem continues, go to Support.
```

To summarize, the failure message must start with "FAILED" in red with **bold** and followed by the detailed error message in new line like above examples. A recovery solution must be provided, e.g. "Use another name and try again." and "Try again later." in above examples.

Bluemix CLI also provides Failed method to print out failure message:

```go
func Run() error {
    ui.Say("Scaling container group '%s'...", terminal.EntityNameColor("xxx"))
    ...
    if err != nil {
        ui.Failed("A server error occurred while scaling the container group.\nTry again later. If the problem continues, go to Support.")
        return err
    }
    ...
}
```

### 2.7. Command Success

When command was successful, the success message should start with "OK" in green with **bold** and followed by the optional details in new line like the following examples:

```
Creating application 'my-app'...
OK
Application 'my-app' was created.
```

The following code snippet can help you print the above message:

```go
ui.Say("Creating application '%s'...",terminal.EntityNameColor("my-app"))
...
ui.Ok()
ui.Say("Application '%s' was created.", terminal.EntityNameColor("my-app"))
```

### 2.8. Warning Message

All of command warnings should be magenta with **bold** like:

```
WARNING:
   Providing your password as a command line option is not recommended
   Your password might be visible to others and might be recorded in your shell history
```

Take the following code as an example for the warning message output:

```go
ui.Warn("WARNING:...")
```

### 2.9. Important Information

The important information displayed to the end-user should be cyan with **bold**. For example:

```
The newer version of Bluemix CLI is available.
You can download it from http://xxx.xxx.xxx
```

The corresponding code snippet:

```go
ui.Say(terminal.PromptColor("The new version of Bluemix CLI ..."))
```

### 2.10. User Input Prompt

Following specifications should be followed to prompt for user input:

1. The input prompt should consist of a prompt message and a right angle bracket '<span style="color: cyan">**>**</span>' with a trailing space.
2. For password prompt, the user input must be hidden.
3. For confirmation prompt, the prompt message should end with `[Y/n]` or `[y/N]` (the capitalized letter indicates the default answer) or `[y/n]` (no default, user input is required).

Following are examples and code snippets:

####Text prompt

```
Logging in to https://api.ng.bluemix.net...
Email>xxx@example.com
Password>
OK
```

Code:

```go
ui.Say("Logging in to https://api.ng.bluemix.net")

var email string
err := ui.Prompt("Email", nil).Resolve(&email)
if err != nil {
    panic(err)
}

var passwd string
err =  ui.Prompt("Password", &terminal.PromptOptions{HideInput: true}).Resolve(&passwd)
if err ! = nil {
    panic(err)
}
...
ui.OK()
```

####Confirmation prompt

```
Are you sure you want to remove the file? [y/N]> y
Removing ...
OK
```

Code:

```go
confirmed := false
err := ui.Prompt("Are you sure you want to remove the file?", nil).Resolve(&confirmed)
if err != nil {
    panic(err)
}

if confirmed {
    ui.Say("Removing...")
    ...
    ui.OK()
}
```

#### Choices prompt

```
Select the plug-in to be upgraded:
1. plugin1
2. plugin2
Enter a number (1)> 2

Upgrading 'plugin2'...
```

Code:

```go
plugins = []string{"plugin1", "plugin2"}
selected := plugins[0] // default is the first plugin

err := ui.ChoicesPrompt("Select the plug-in to be upgraded:", plugins, nil).Resolve(&selected)
if err != nil {
    panic(err)
}

ui.Say("upgrading '%s'...", terminal.EntityNameColor(selected))
```

### 2.11. Table Output

For consistent user experience, developers of Bluemix CLI plugin should comply with the following table output specifications:

1. Table header should be bold.
2. Each word in table header should be capitalized.
3. The entity name or resource name in table (usually the first column) should be cyan with bold.

The following is an example:

```
Name       Description
my-app     this is my application.
demo-app   This is a long long long ...
           description.
```           

Using APIs provided by Bluemix CLI can let you easily print results in table format:
```go
import "github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/terminal"

func (demo *DemoPlugin) PrintTable() {
    ui := terminal.NewStdUI()

    table := ui.Table([]string{"Name", "Description"})
    table.Add("my-app", "Description of my app.")
    table.Add("demo-app", "This is a long long long ...\ndescription.")
    table.Print()
}
```

## 3. Tracing

Bluemix CLI provides utility for tracing based on "BLUEMIX\_TRACE" environment variable. The trace will be disabled if environment variable "BLUEMIX\_TRACE" was not set or it was set to "false" (case ignored), which means, in that case, the invocation of trace API has no effect. If "BLUEMIX\_TRACE" was set to "true" (case ignored), the trace will be printed on the terminal. Otherwise, the value of "BLUEMIX\_TRACE" will be treated as the path of trace file.

It's pretty much easy on using Bluemix CLI trace API.
```go
import "github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/trace"

func (demo *DemoPlugin) Run(context plugin.PluginContext, args []string) {
    // first, initialize the trace logger
    trace.Logger = trace.NewLogger(context.Trace())

    // start using the trace logger
    trace.Logger.Println("Start to initialize Demo plugin.")
    ...
    trace.Logger.Printf("%s plugin initialized.", "Demo")
}
```

## 4. HTTP Utilities

### 4.1. HTTP tracing

TraceLoggingTransport in package 'bluemix-cli-sdk/bluemix/http' is a thin wrapper around HTTP transport. It dumps each HTTP request and its response by using the trace logger.

First, you need to initialize the trace logger.
```go
trace.Logger = trace.NewLogger(pluginContext.Trace())
```

Then you can create a HTTP client with TraceLoggingTransport to send the request:
```go
import (
    "github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/http"
    gohttp "net/http"
)

client := &gohttp.Client{
    Transport: http.NewTraceLoggingTransport(gohttp.DefaultTransport)
}
client.Get("http://www.example.com")
```
Then during each round-trip, the trace logger dumps the request and its response.

### 4.2. REST client

HTTP interaction with a remote server is a common task in both core and plugin commands. Package “bluemix-cli-sdk/common/rest” provides APIs for building a REST API request and a REST client for sending the request.

For example, to create a GET request:
```go
import "github.ibm.com/Bluemix/bluemix-cli-sdk/common/rest"

var r = rest.GetRequest("http://www.example.com")
// equal to
// var r = NewRequest(“http://www.example.com”).Method(“GET”)
```

To add a query string in the URL:
```go
r.Query("foo", "bar")
```

To add or set HTTP headers in the request:
```go
r.Add("Accept-Encoding","gzip")
r.Set("Accept", "application/json")
```

To create a POST request, and send form data:
```go
var r = rest.PostRequest("http://www.example.com")
r.Field("foo", "bar")
```

And to upload a file:
```go
f, err := os.Open("1.img")
if err != nil {
   // handle error
}
r.File("file1", rest.File{
    Name: f.Name(),
    Content: f,
    Type: "image/jpeg", // Optional. Default is "application/octet-stream"
})
```

You can send form data and upload multiple files in a same request.

To post a JSON data in the request, you can simply pass a Go struct to the Body () method. The method automatically encodes the Go struct to a JSON string.
For example:
```go
type Foo struct
    Name string
}
var r = rest.PostRequest("http://www.example.com")
r.Body(Foo{Name: "bar"})
```

The Body() method also supports raw string and steam. The above example can also be written as:
```go
r.Body("{\"name\": \"bar\"}")
// Or
r.Body(strings.NewReader("{\"name\": \"foo1\"}"))
```

After the request is created, you can then create a REST client to send the request. The REST client is safe for concurrent use by multiple Go routines and thus is recommend to be reused.

To create a REST client:
```go
client := rest.NewClient()
```

By default, the client use Go’s standard HTTP client. You can override it:
```go
client.HTTPClient = &http.Client{
    Timeout: 60 * time.Second,
}
```

Also, you can set default HTTP header for all outgoing requests:
```go
h := http.Header{}
h.Set("User-Agent", "Bluemix CLI")
Client.DefaultHeader = h
```

Now, you can invoke client’s Do() method to send the request. The method automatically unmarshals the response body to the Go struct. If server response’s status code is 2xx, successV is unmarshaled; otherwise, errorV is un- marshaled if exists. If errorV is not provided or not successfully unmarshaled, an ErrorResponse typed error is returned which has status code and response text.
```go
var successV Foo
var errorV = struct {
    Message string
}{}

resp, err := client.Do(r, &successV, &errorV)
if errorV != nil || err != nil {
    // handle error
}
```

## 5. Utility for Unit Testing

We highly recommended that terminal.StdUI was used in your code for output, because it can be replaced by FakeUI which is a utility provided by Bluemix CLI for easy unit testing.

FakeUI can intercept all of output and store the output in a buffer temporarily, so that you can compare the actual output and expected output easily. Take the following code snippet as an example:

```go
// Here is the source code to be tested:
import (
    "github.ibm.com/Bluemix/bluemix-cli-sdk/plugin"
    "github.ibm.com/Bluemix/bluemix-cli-sdk/bluemix/terminal"
)

type DemoPluginstruct {}

func (demo *DemoPlugin) Run(context plugin.PluginContext, args []string){
    ...
    if args[0] == "start" {
        demo.Start(terminal.StdUI)
    }
}

func (demo *DemoPlugin) Start(ui terminal.UI){
    ...
    ui.echo("OK")
    ...
}

// The following is the testing code:
import "github.ibm.com/Bluemix/bluemix-cli-sdk/testhelpers/terminal"

func TestStart() {
    demoPlugin := DemoPlugin{}
    fakeUI := terminal.NewFakeUI()
    demoPlugin.Start(fakeUI)

    // Now you can use any third-party unit testing framework to assert the result,
    // for example:
    assert.IsTrue(fakeUI.ContainsOutput("OK"))
    assert.IsFalse(fakeUI.ContainsOutput("FAILED"))
}
```

## 6. Globalization

Bluemix CLI tends to be used globally. Both Bluemix CLI and its plugins should support globalization. We have enabled internalization (i18n) for CLI's base commands with the help of the third-party tool "[go-i18n](https://github.com/nicksnyder/go-i18n)". To keep user experience consistent, we recommend plugin developers follow CLI's way of i18n enablement.

Here's the workflow:

_Step 1:_ Add new strings or replace existing strings with `T()` function calls to load the translated strings.  For example:

```go
T("Installing the plugin...")

//With variable substitution
T("Plugin '{{.Name}}' was successfully installed.",
map[string]interface{}{"Name": pluginName})
```

Note: T is type of [TranslateFunc](https://godoc.org/github.com/nicksnyder/go-i18n/i18n#TranslateFunc) which is set in i18n initialization (see #4).

_Step 2:_ Prepare translation files. Add all strings in en-us.all.json, and then use go-i18n CLI to generate translation files for other languages. A sample en-us.all.json is like:

```javascript
[
   {
          "id": "Installing the plugin…",
          "translation": "Installing the plugin…"
   },
   {
          "id": "Plugin '{{.Name}}' was successfully installed.",
          "translation": " Plugin '{{.Name}}' was successfully installed."
   },
   ...
]
```

To generate translation files for other language, e.g, zh\_Hans (Simplified Chinese) and fr\_BR, create empty files zh-hans.all.json and fr-br.all.json, and run:

```bash
goi18n –outputdir <directory\_of\_generated\_translation\_files> en-us.all.json zh-hans.all.json fr-br.all.json
```

The above command will generate 2 output files for each language: xx-yy.all.json contains all strings for the language, and xx-yy.untranslated.json contains untranslated strings. After the strings are translated, they should be merged back into xx-yy.all.json. For more details, refer to goi18n CLI's help by 'goi18n –help'.

_Step 3:_ Package translation files.
Bluemix CLI is to be built as a stand-alone binary distribution. In order to load i18n resource files in code, we use [go-bindata](https://github.com/jteeuwen/go-bindata) to auto-generate Go source code from all i18n resource files and the compile them into the binary. You can write a script to do it automatically during build. A sample script could be like:

```bash
#!/bin/bash

set -e

go get github.com/jteeuwen/go-bindata/...

echo "Generating i18n resource file ..."
$GOPATH/bin/go-bindata -pkg resources -o bluemix/resources/i18n\_resources.go bluemix/i18n/resources
```

After execution, a source file 'bluemix/resources/i18n\_resources.go' with package name 'resources' is generated to embed all files under bluemix/i18n/resource. To access a resource file, invoke resources.Asset("bluemix/i18n/resources/<fileName>") which returns bytes of the file.

_Step 4:_ Initialize i18n
    `T()` must be initialized before use. During i18n initialization in Bluemix CLI, user locale is used if it's set in ~/.bluemix/config.json ( plugin can get user locale via PluginContext.Locale() ); otherwise, system locale is auto discovered (see [jibber\_jabber](https://github.com/cloudfoundry/jibber_jabber)) and used. If system locale is not detected or supported, default locale en\_US is then used. Next, we initialize the translate function with the locale. A sample code could be like:

```go
func initWithLocale(locale string) goi18n.TranslateFunc {
    err := loadFromAsset(locale)
    if err != nil {
        panic(err)
    }
    return goi18n.MustTfunc(locale, DEFAULT_LOCALE)
}

// load translation asset for the given locale
func loadFromAsset(locale string) (err error) {
    assetName := locale + ".all.json"
    assetKey := filepath.Join(resourcePath, assetName)
    bytes, err := resources.Asset(assetKey)
    if err != nil {
       return
    }
    err = goi18n.ParseTranslationFileBytes(assetName, bytes)
    return
}
```
