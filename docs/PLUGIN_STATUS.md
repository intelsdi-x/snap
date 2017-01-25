# Plugin Status

Not all plugins are designed with the same level of readiness in mind. To visualize that fact while also allowing everyone to share their work, we have introduced the concept of "tiers" of plugins. These tiers are communicated as statuses which will be visualized everywhere you see the [Plugin Catalog](PLUGIN_CATALOG.md).

Note that plugins can and likely will be demoted if they fall significantly behind best practices of the project. I know that's vague, but it will have to do for this first version :grimacing:

### What this is

* A way to quickly see the highest quality plugins
* A way to encourage contribution to existing plugins instead of forking

### What this is not

* Not a definition of the plugin version - that's done through `releases`
* Not a contract of support - we do our best as maintainers of Snap, but make no promises

## Plugin Status Matrix

All plugins meet a set of minimum requirements to be included at its status level. Note that some earlier stage plugins *can* have other requirements implemented, but they *must* have all checkboxes to be in the next tier.

| Requirements                   |     Unlabeled      |    Experimental    |      Approved      |     Supported      |
|:-------------------------------|:------------------:|:------------------:|:------------------:|:------------------:|
| No naming conflict             | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Includes README.md             |                    | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Active maintainers             |                    | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Release includes binary        |                    |                    | :white_check_mark: | :white_check_mark: |
| Known to compile               |                    |                    | :white_check_mark: | :white_check_mark: |
| Known to load in Snap          |                    |                    | :white_check_mark: | :white_check_mark: |
| Uses Snap plugin library       |                    |                    | :white_check_mark: | :white_check_mark: |
| Reasonable test coverage       |                    |                    | :white_check_mark: | :white_check_mark: |
| Includes example tasks         |                    |                    | :white_check_mark: | :white_check_mark: |
| Includes dependency management |                    |                    | :white_check_mark: | :white_check_mark: |
| Includes CI status             |                    |                    | :white_check_mark: | :white_check_mark: |
| Includes license               |                    |                    | :white_check_mark: | :white_check_mark: |
| Support provided by a company  |                    |                    |                    | :white_check_mark: |

## Supported Plugins

These are our premier plugins for the Snap telemetry framework and its users. These are designed to follow all of our recommended practices. Issues are also closely monitored by its supporting company ([read more about that here](#more-on-support-for-plugins)). While companies may support some plugins they contribute to Snap, not all contributions will meet the Supported standard.

We prefer to not have other repositories that overlap with Supported plugins and suggest contributing to the existing version to help keep this list small and effective. Like all plugins, community contribution is welcome.


## Approved Plugins

These are primarily community-contributed plugins that meet or exceed the project's best practices. These plugins have been vetted by Snap maintainers as of the date listed in the Plugin Catalog. They are excellent references and are quite likely ready for use in your own Snap deployment.

We prefer to not have other repositories that overlap with Approved plugins and suggest contributing to the existing version to help keep this list small and effective. Don't be shy about reaching out to existing plugin authors to see if you can help improve upon it.

## Experimental

These plugins are in development and are not yet complete, but are shared with the community for feedback and testing. They are shared with the community for feedback and testing. Think of this tier as an incubation phase that should move toward Approved or drop down to Unlabeled.

We prefer to not have other repositories that overlap with Experimental plugins and suggest contributing to the existing version to help keep this list small and effective. Don't be shy about reaching out to existing plugin authors to see if you can help improve upon it.


## All Other Plugins (Unlabeled)

These are plugins in varying phases of completeness and are shared for reference. They do not necessarily follow best practices for plugin development. We welcome anyone forking these plugins and working toward Approved status.

## Changing Tiers 

If you find a plugin that should move between tiers (ex. from Unlabeled to Approved, Supported to Experimental or Approved to Experimental), open an issue to do so on the [main Snap repository](https://github.com/intelsdi-x/snap/issues). Please include any corresponding issues opened on the plugin repository as well. 

## More On Support For Plugins

Snap is an open source project originated and actively maintained by Intel with the goal of becoming a broad community standard for telemetry. We stand behind Supported plugins with the intention of making them part of your monitoring infrastructure. As other companies adopt Snap as their standard telemetry framework, they may also choose to support various plugins. We want to make sure you know which ones, and what companies, are here to support you. 