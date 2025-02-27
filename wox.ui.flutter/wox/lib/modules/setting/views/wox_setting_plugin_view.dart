import 'dart:convert';

import 'package:dynamic_tabbar/dynamic_tabbar.dart' as dt;
import 'package:fluent_ui/fluent_ui.dart';
import 'package:flutter/material.dart' as material;
import 'package:flutter/services.dart';
import 'package:flutter_image_slideshow/flutter_image_slideshow.dart';
import 'package:get/get.dart';
import 'package:wox/components/plugin/wox_setting_plugin_head_view.dart';
import 'package:wox/components/plugin/wox_setting_plugin_label_view.dart';
import 'package:wox/components/plugin/wox_setting_plugin_newline_view.dart';
import 'package:wox/components/plugin/wox_setting_plugin_select_ai_model_view.dart';
import 'package:wox/components/plugin/wox_setting_plugin_select_view.dart';
import 'package:wox/components/plugin/wox_setting_plugin_table_view.dart';
import 'package:wox/components/wox_image_view.dart';
import 'package:wox/entity/setting/wox_plugin_setting_label.dart';
import 'package:wox/entity/setting/wox_plugin_setting_select_ai_model.dart';
import 'package:wox/entity/setting/wox_plugin_setting_table.dart';
import 'package:wox/entity/wox_plugin.dart';
import 'package:wox/entity/setting/wox_plugin_setting_checkbox.dart';
import 'package:wox/entity/setting/wox_plugin_setting_head.dart';
import 'package:wox/entity/setting/wox_plugin_setting_newline.dart';
import 'package:wox/entity/setting/wox_plugin_setting_select.dart';
import 'package:wox/entity/setting/wox_plugin_setting_textbox.dart';
import 'package:wox/components/plugin/wox_setting_plugin_checkbox_view.dart';
import 'package:wox/components/plugin/wox_setting_plugin_textbox_view.dart';
import 'package:wox/modules/setting/wox_setting_controller.dart';
import 'package:wox/utils/colors.dart';

class WoxSettingPluginView extends GetView<WoxSettingController> {
  const WoxSettingPluginView({super.key});

  Widget pluginList() {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.only(bottom: 20),
          child: Focus(
            autofocus: true,
            onKeyEvent: (FocusNode node, KeyEvent event) {
              if (event is KeyDownEvent) {
                switch (event.logicalKey) {
                  case LogicalKeyboardKey.escape:
                    controller.hideWindow();
                    return KeyEventResult.handled;
                }
              }

              return KeyEventResult.ignored;
            },
            child: Obx(() {
              return TextBox(
                autofocus: true,
                controller: controller.filterPluginKeywordController,
                placeholder: 'Search ${controller.filteredPluginDetails.length} plugins',
                padding: const EdgeInsets.all(10),
                suffix: const Padding(
                  padding: EdgeInsets.only(right: 8.0),
                  child: Icon(FluentIcons.search),
                ),
                onChanged: (value) {
                  controller.filterPlugins();
                  controller.setFirstFilteredPluginDetailActive();
                },
              );
            }),
          ),
        ),
        Expanded(
          child: Scrollbar(
            thumbVisibility: false,
            child: Obx(() {
              return ListView.builder(
                primary: true,
                itemCount: controller.filteredPluginDetails.length,
                itemBuilder: (context, index) {
                  final plugin = controller.filteredPluginDetails[index];
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 8.0),
                    child: Obx(() {
                      final isActive = controller.activePluginDetail.value.id == plugin.id;
                      return Container(
                        decoration: BoxDecoration(
                          color: isActive ? SettingPrimaryColor : Colors.transparent,
                          borderRadius: BorderRadius.circular(4),
                        ),
                        child: GestureDetector(
                          behavior: HitTestBehavior.translucent,
                          onTap: () {
                            controller.activePluginDetail.value = plugin;
                          },
                          // fluent listTile is not clickable
                          child: material.ListTile(
                            contentPadding: const EdgeInsets.only(left: 6, right: 6),
                            leading: WoxImageView(woxImage: plugin.icon, width: 32),
                            title: Text(plugin.name,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: TextStyle(
                                  fontSize: 15,
                                  color: isActive ? Colors.white : Colors.black,
                                )),
                            subtitle: Row(
                              mainAxisAlignment: MainAxisAlignment.start,
                              children: [
                                Text(
                                  plugin.version,
                                  maxLines: 1, // Limiting the description to two lines
                                  overflow: TextOverflow.ellipsis, // Add ellipsis for overflow
                                  style: TextStyle(
                                    color: isActive ? Colors.white : Colors.grey,
                                    fontSize: 12,
                                  ),
                                ),
                                const SizedBox(width: 10),
                                Text(
                                  plugin.author,
                                  maxLines: 1, // Limiting the description to two lines
                                  overflow: TextOverflow.ellipsis, // Add ellipsis for overflow
                                  style: TextStyle(
                                    color: isActive ? Colors.white : Colors.grey,
                                    fontSize: 12,
                                  ),
                                ),
                              ],
                            ),
                            trailing: pluginTrailIcon(plugin, isActive),
                          ),
                        ),
                      );
                    }),
                  );
                },
              );
            }),
          ),
        ),
      ],
    );
  }

  Widget pluginTrailIcon(PluginDetail plugin, bool isActive) {
    if (controller.isStorePluginList.value) {
      if (plugin.isInstalled) {
        return Icon(FluentIcons.skype_circle_check, color: isActive ? Colors.white : Colors.green);
      }
    }
    return const SizedBox();
  }

  Widget pluginDetail() {
    return Expanded(
      child: Obx(() {
        final plugin = controller.activePluginDetail.value;
        return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Padding(
            padding: const EdgeInsets.only(bottom: 8.0, left: 10),
            child: Row(
              children: [
                WoxImageView(woxImage: plugin.icon, width: 32),
                Padding(
                  padding: const EdgeInsets.only(left: 8.0),
                  child: Text(
                    plugin.name,
                    style: const TextStyle(
                      fontSize: 20,
                    ),
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.only(left: 10.0),
                  child: Text(
                    plugin.version,
                    style: const TextStyle(
                      color: Colors.grey,
                    ),
                  ),
                ),
                if (plugin.isDev)
                  // dev tag, warning color with warning border
                  Padding(
                    padding: const EdgeInsets.only(left: 10.0),
                    child: Container(
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: SettingWarningColor,
                        border: Border.all(color: SettingWarningColor),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: const Text(
                        'dev',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 12,
                        ),
                      ),
                    ),
                  ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(bottom: 8.0, left: 16),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  plugin.author,
                  style: const TextStyle(
                    color: Colors.grey,
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.only(left: 18.0),
                  child: HyperlinkButton(
                    onPressed: () {
                      controller.openPluginWebsite(plugin.website);
                    },
                    child: Row(
                      children: [
                        Text(
                          "website",
                          style: TextStyle(
                            color: Colors.blue,
                          ),
                        ),
                        Padding(
                          padding: const EdgeInsets.only(left: 4.0),
                          child: Icon(
                            FluentIcons.open_in_new_tab,
                            size: 12,
                            color: Colors.blue,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(bottom: 8.0, left: 16),
            child: Row(
              children: [
                if (plugin.isInstalled && !plugin.isSystem)
                  Padding(
                    padding: const EdgeInsets.only(right: 8.0),
                    child: Button(
                      onPressed: () {
                        controller.uninstallPlugin(plugin);
                      },
                      child: const Text('Uninstall'),
                    ),
                  ),
                if (!plugin.isInstalled)
                  Padding(
                    padding: const EdgeInsets.only(right: 8.0),
                    child: Button(
                      onPressed: () {
                        controller.installPlugin(plugin);
                      },
                      child: const Text('Install'),
                    ),
                  ),
                if (plugin.isInstalled && !plugin.isDisable)
                  Padding(
                    padding: const EdgeInsets.only(right: 8.0),
                    child: Button(
                      onPressed: () {
                        controller.disablePlugin(plugin);
                      },
                      child: const Text('Disable'),
                    ),
                  ),
                if (plugin.isInstalled && plugin.isDisable)
                  Padding(
                    padding: const EdgeInsets.only(right: 8.0),
                    child: Button(
                      onPressed: () {
                        controller.enablePlugin(plugin);
                      },
                      child: const Text('Enable'),
                    ),
                  ),
              ],
            ),
          ),
          Expanded(
            child: dt.DynamicTabBarWidget(
              isScrollable: true,
              showBackIcon: false,
              showNextIcon: false,
              tabAlignment: material.TabAlignment.start,
              onAddTabMoveTo: dt.MoveToTab.idol,
              labelColor: SettingPrimaryColor,
              indicatorColor: SettingPrimaryColor,
              dynamicTabs: [
                if (controller.shouldShowSettingTab())
                  dt.TabData(
                    index: 0,
                    title: const material.Tab(
                      child: Text('Settings'),
                    ),
                    content: pluginTabSetting(),
                  ),
                dt.TabData(
                  index: 1,
                  title: const material.Tab(
                    child: Text('Trigger Keywords'),
                  ),
                  content: pluginTabTriggerKeywords(),
                ),
                dt.TabData(
                  index: 2,
                  title: const material.Tab(
                    child: Text('Commands'),
                  ),
                  content: pluginTabCommand(),
                ),
                dt.TabData(
                  index: 3,
                  title: const material.Tab(
                    child: Text('Description'),
                  ),
                  content: pluginTabDescription(),
                ),
                dt.TabData(
                  index: 4,
                  title: const material.Tab(
                    child: Text('Privacy'),
                  ),
                  content: pluginTabPrivacy(),
                ),
              ],
              onTabControllerUpdated: (tabController) {
                controller.activePluginTabController = tabController;
              },
              onTabChanged: (index) {},
            ),
          ),
        ]);
      }),
    );
  }

  Widget pluginTabDescription() {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            controller.activePluginDetail.value.description,
          ),
          const SizedBox(height: 20),
          controller.activePluginDetail.value.screenshotUrls.isNotEmpty
              ? ImageSlideshow(
                  width: double.infinity,
                  height: 400,
                  indicatorColor: SettingPrimaryColor,
                  children: [
                    ...controller.activePluginDetail.value.screenshotUrls.map((e) => Image.network(e)),
                  ],
                )
              : const SizedBox(),
        ],
      ),
    );
  }

  Widget pluginTabSetting() {
    return Obx(() {
      var plugin = controller.activePluginDetail.value;
      return Padding(
        padding: const EdgeInsets.all(16.0),
        child: SingleChildScrollView(
          child: Wrap(
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              ...plugin.settingDefinitions.map(
                (e) {
                  if (e.type == "checkbox") {
                    return WoxSettingPluginCheckbox(
                      value: plugin.setting.settings[e.value.key] ?? "",
                      item: e.value as PluginSettingValueCheckBox,
                      onUpdate: (key, value) async {
                        await controller.updatePluginSetting(plugin.id, key, value);
                        controller.refreshPluginList();
                      },
                    );
                  }
                  if (e.type == "textbox") {
                    return WoxSettingPluginTextBox(
                      value: plugin.setting.settings[e.value.key] ?? "",
                      item: e.value as PluginSettingValueTextBox,
                      onUpdate: (key, value) async {
                        await controller.updatePluginSetting(plugin.id, key, value);
                        controller.refreshPluginList();
                      },
                    );
                  }
                  if (e.type == "newline") {
                    return WoxSettingPluginNewLine(
                      value: "",
                      item: e.value as PluginSettingValueNewLine,
                      onUpdate: (key, value) async {
                        await controller.updatePluginSetting(plugin.id, key, value);
                        controller.refreshPluginList();
                      },
                    );
                  }
                  if (e.type == "select") {
                    return WoxSettingPluginSelect(
                      value: plugin.setting.settings[e.value.key] ?? "",
                      item: e.value as PluginSettingValueSelect,
                      onUpdate: (key, value) async {
                        await controller.updatePluginSetting(plugin.id, key, value);
                        controller.refreshPluginList();
                      },
                    );
                  }
                  if (e.type == "selectAIModel") {
                    return WoxSettingPluginSelectAIModel(
                      value: plugin.setting.settings[e.value.key] ?? "",
                      item: e.value as PluginSettingValueSelectAIModel,
                      onUpdate: (key, value) async {
                        await controller.updatePluginSetting(plugin.id, key, value);
                        controller.refreshPluginList();
                      },
                    );
                  }
                  if (e.type == "head") {
                    return WoxSettingPluginHead(
                      value: plugin.setting.settings[e.value.key] ?? "",
                      item: e.value as PluginSettingValueHead,
                      onUpdate: (key, value) async {
                        await controller.updatePluginSetting(plugin.id, key, value);
                        controller.refreshPluginList();
                      },
                    );
                  }
                  if (e.type == "label") {
                    return WoxSettingPluginLabel(
                      value: plugin.setting.settings[e.value.key] ?? "",
                      item: e.value as PluginSettingValueLabel,
                      onUpdate: (key, value) async {
                        await controller.updatePluginSetting(plugin.id, key, value);
                        controller.refreshPluginList();
                      },
                    );
                  }
                  if (e.type == "table") {
                    return WoxSettingPluginTable(
                      value: plugin.setting.settings[e.value.key] ?? "",
                      item: e.value as PluginSettingValueTable,
                      onUpdate: (key, value) async {
                        await controller.updatePluginSetting(plugin.id, key, value);
                        controller.refreshPluginList();
                      },
                    );
                  }

                  return Text(e.type);
                },
              )
            ],
          ),
        ),
      );
    });
  }

  Widget pluginTabTriggerKeywords() {
    var plugin = controller.activePluginDetail.value;
    if (plugin.triggerKeywords.isEmpty) {
      return const Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('This plugin has no trigger keywords'),
          ],
        ),
      );
    }

    return WoxSettingPluginTable(
      value: json.encode(plugin.triggerKeywords.map((e) => {"keyword": e}).toList()),
      tableWidth: 680,
      item: PluginSettingValueTable.fromJson({
        "Key": "_triggerKeywords",
        "Columns": [
          {
            "Key": "keyword",
            "Label": "Keyword",
            "Tooltip": "The keyword that triggers this plugin. * presents no keyword, any input will trigger this plugin.",
            "Type": "text",
            "TextMaxLines": 1,
            "Validators": [
              {"Type": "not_empty"}
            ],
          },
        ],
        "SortColumnKey": "keyword"
      }),
      onUpdate: (key, value) {
        final List<String> triggerKeywords = [];
        for (var item in json.decode(value)) {
          triggerKeywords.add(item["keyword"]);
        }
        plugin.triggerKeywords = triggerKeywords;
        controller.updatePluginTriggerKeywords(plugin.id, triggerKeywords);
        controller.refreshPluginList();
      },
    );
  }

  Widget pluginTabCommand() {
    var plugin = controller.activePluginDetail.value;
    if (plugin.commands.isEmpty) {
      return const Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('This plugin has no command'),
          ],
        ),
      );
    }

    return WoxSettingPluginTable(
      value: json.encode(plugin.commands),
      tableWidth: 680,
      readonly: true,
      item: PluginSettingValueTable.fromJson({
        "Key": "_commands",
        "Columns": [
          {
            "Key": "Command",
            "Label": "Name",
            "Width": 120,
            "Type": "text",
            "TextMaxLines": 1,
            "Validators": [
              {"Type": "not_empty"}
            ],
          },
          {
            "Key": "Description",
            "Label": "Description",
            "Type": "text",
            "TextMaxLines": 1,
            "Validators": [
              {"Type": "not_empty"}
            ],
          }
        ],
        "SortColumnKey": "Command"
      }),
      onUpdate: (key, value) {},
    );
  }

  Widget pluginTabPrivacy() {
    var plugin = controller.activePluginDetail.value;
    var noDataAccess = const Padding(
      padding: EdgeInsets.all(16),
      child: Text('This plugin requires no data access'),
    );

    if (plugin.features.isEmpty) {
      return noDataAccess;
    }

    List<String> params = [];

    //check if "queryEnv" feature is exist and list it's params
    var queryEnv = plugin.features.where((element) => element.name == "queryEnv").toList();
    if (queryEnv.isNotEmpty) {
      queryEnv.first.params.forEach((key, value) {
        if (value == "true") {
          params.add(key);
        }
      });
    }

    // check if llmChat feature is exist
    var llmChat = plugin.features.where((element) => element.name == "llm").toList();
    if (llmChat.isNotEmpty) {
      params.add("llm");
    }

    if (params.isEmpty) {
      return noDataAccess;
    }

    return Padding(
      padding: const EdgeInsets.all(16.0),
      child: Expanded(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('This plugin requires access to the following data:'),
            ...params.map((e) {
              if (e == "requireActiveWindowName") {
                return privacyItem(
                  material.Icons.window,
                  'Active window name',
                  'E.g. you are using google chrome to view webpages, you activate Wox and this plugin will get the active window name as "Google Chrome"',
                );
              }
              if (e == "requireActiveWindowPid") {
                return privacyItem(
                  material.Icons.window,
                  'Active window process id',
                  'E.g. you are using google chrome to view webpages, you activate Wox and this plugin will get the active window process id as "1234"',
                );
              }
              if (e == "requireActiveBrowserUrl") {
                return privacyItem(
                  material.Icons.web_sharp,
                  'Active browser URL',
                  'E.g. you are using google chrome to view webpages, you activate Wox and this plugin will get the url of active tab you are viewing',
                );
              }
              if (e == "llm") {
                return privacyItem(
                  material.Icons.chat,
                  'Large Language Model (LLM)',
                  'This plugin uses large language model to provide better results, you need to configure the model in LLM Tools plugin first',
                );
              }
              return Text(e);
            }),
          ],
        ),
      ),
    );
  }

  Widget privacyItem(IconData icon, String title, String description) {
    return Padding(
      padding: const EdgeInsets.only(top: 20.0),
      child: Column(
        children: [
          Row(
            children: [
              Icon(icon),
              SizedBox(width: 10),
              Text(title),
            ],
          ),
          const SizedBox(height: 6),
          Row(
            children: [
              const SizedBox(width: 30),
              Flexible(
                child: Text(
                  description,
                  style: TextStyle(
                    color: Colors.grey[100],
                  ),
                ),
              ),
            ],
          )
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(20),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 250,
            child: pluginList(),
          ),
          // This is your divider
          Container(
            width: 1,
            height: double.infinity,
            color: Colors.grey[30],
            margin: const EdgeInsets.only(right: 10, left: 10),
          ),
          pluginDetail(),
        ],
      ),
    );
  }
}
