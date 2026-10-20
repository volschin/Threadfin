var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __commonJS = (cb, mod) => function __require() {
  try {
    return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
  } catch (e) {
    throw mod = 0, e;
  }
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
  // If the importer is in node compatibility mode or this is not an ESM
  // file that has been converted to a CommonJS file using a Babel-
  // compatible transform (i.e. "__esModule" has not been set), then set
  // "default" to the CommonJS "module.exports" for node compatibility.
  isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
  mod
));

// node_modules/clipboard/dist/clipboard.js
var require_clipboard = __commonJS({
  "node_modules/clipboard/dist/clipboard.js"(exports, module) {
    (function webpackUniversalModuleDefinition(root, factory) {
      if (typeof exports === "object" && typeof module === "object")
        module.exports = factory();
      else if (typeof define === "function" && define.amd)
        define([], factory);
      else if (typeof exports === "object")
        exports["ClipboardJS"] = factory();
      else
        root["ClipboardJS"] = factory();
    })(exports, function() {
      return (
        /******/
        (function() {
          var __webpack_modules__ = {
            /***/
            686: (
              /***/
              (function(__unused_webpack_module, __webpack_exports__, __webpack_require__2) {
                "use strict";
                __webpack_require__2.d(__webpack_exports__, {
                  "default": function() {
                    return (
                      /* binding */
                      clipboard2
                    );
                  }
                });
                var tiny_emitter = __webpack_require__2(279);
                var tiny_emitter_default = /* @__PURE__ */ __webpack_require__2.n(tiny_emitter);
                var listen = __webpack_require__2(370);
                var listen_default = /* @__PURE__ */ __webpack_require__2.n(listen);
                var src_select = __webpack_require__2(817);
                var select_default = /* @__PURE__ */ __webpack_require__2.n(src_select);
                ;
                function command(type) {
                  try {
                    return document.execCommand(type);
                  } catch (err) {
                    return false;
                  }
                }
                ;
                var ClipboardActionCut = function ClipboardActionCut2(target) {
                  var selectedText = select_default()(target);
                  command("cut");
                  return selectedText;
                };
                var actions_cut = ClipboardActionCut;
                ;
                function createFakeElement(value) {
                  var isRTL3 = document.documentElement.getAttribute("dir") === "rtl";
                  var fakeElement = document.createElement("textarea");
                  fakeElement.style.fontSize = "12pt";
                  fakeElement.style.border = "0";
                  fakeElement.style.padding = "0";
                  fakeElement.style.margin = "0";
                  fakeElement.style.position = "absolute";
                  fakeElement.style[isRTL3 ? "right" : "left"] = "-9999px";
                  var yPosition = window.pageYOffset || document.documentElement.scrollTop;
                  fakeElement.style.top = "".concat(yPosition, "px");
                  fakeElement.setAttribute("readonly", "");
                  fakeElement.value = value;
                  return fakeElement;
                }
                ;
                var fakeCopyAction = function fakeCopyAction2(value, options) {
                  var fakeElement = createFakeElement(value);
                  options.container.appendChild(fakeElement);
                  var selectedText = select_default()(fakeElement);
                  command("copy");
                  fakeElement.remove();
                  return selectedText;
                };
                var ClipboardActionCopy = function ClipboardActionCopy2(target) {
                  var options = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {
                    container: document.body
                  };
                  var selectedText = "";
                  if (typeof target === "string") {
                    selectedText = fakeCopyAction(target, options);
                  } else if (target instanceof HTMLInputElement && !["text", "search", "url", "tel", "password"].includes(target === null || target === void 0 ? void 0 : target.type)) {
                    selectedText = fakeCopyAction(target.value, options);
                  } else {
                    selectedText = select_default()(target);
                    command("copy");
                  }
                  return selectedText;
                };
                var actions_copy = ClipboardActionCopy;
                ;
                function _typeof(obj) {
                  "@babel/helpers - typeof";
                  if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") {
                    _typeof = function _typeof2(obj2) {
                      return typeof obj2;
                    };
                  } else {
                    _typeof = function _typeof2(obj2) {
                      return obj2 && typeof Symbol === "function" && obj2.constructor === Symbol && obj2 !== Symbol.prototype ? "symbol" : typeof obj2;
                    };
                  }
                  return _typeof(obj);
                }
                var ClipboardActionDefault = function ClipboardActionDefault2() {
                  var options = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
                  var _options$action = options.action, action = _options$action === void 0 ? "copy" : _options$action, container = options.container, target = options.target, text = options.text;
                  if (action !== "copy" && action !== "cut") {
                    throw new Error('Invalid "action" value, use either "copy" or "cut"');
                  }
                  if (target !== void 0) {
                    if (target && _typeof(target) === "object" && target.nodeType === 1) {
                      if (action === "copy" && target.hasAttribute("disabled")) {
                        throw new Error('Invalid "target" attribute. Please use "readonly" instead of "disabled" attribute');
                      }
                      if (action === "cut" && (target.hasAttribute("readonly") || target.hasAttribute("disabled"))) {
                        throw new Error(`Invalid "target" attribute. You can't cut text from elements with "readonly" or "disabled" attributes`);
                      }
                    } else {
                      throw new Error('Invalid "target" value, use a valid Element');
                    }
                  }
                  if (text) {
                    return actions_copy(text, {
                      container
                    });
                  }
                  if (target) {
                    return action === "cut" ? actions_cut(target) : actions_copy(target, {
                      container
                    });
                  }
                };
                var actions_default = ClipboardActionDefault;
                ;
                function clipboard_typeof(obj) {
                  "@babel/helpers - typeof";
                  if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") {
                    clipboard_typeof = function _typeof2(obj2) {
                      return typeof obj2;
                    };
                  } else {
                    clipboard_typeof = function _typeof2(obj2) {
                      return obj2 && typeof Symbol === "function" && obj2.constructor === Symbol && obj2 !== Symbol.prototype ? "symbol" : typeof obj2;
                    };
                  }
                  return clipboard_typeof(obj);
                }
                function _classCallCheck(instance, Constructor) {
                  if (!(instance instanceof Constructor)) {
                    throw new TypeError("Cannot call a class as a function");
                  }
                }
                function _defineProperties(target, props) {
                  for (var i = 0; i < props.length; i++) {
                    var descriptor = props[i];
                    descriptor.enumerable = descriptor.enumerable || false;
                    descriptor.configurable = true;
                    if ("value" in descriptor) descriptor.writable = true;
                    Object.defineProperty(target, descriptor.key, descriptor);
                  }
                }
                function _createClass(Constructor, protoProps, staticProps) {
                  if (protoProps) _defineProperties(Constructor.prototype, protoProps);
                  if (staticProps) _defineProperties(Constructor, staticProps);
                  return Constructor;
                }
                function _inherits(subClass, superClass) {
                  if (typeof superClass !== "function" && superClass !== null) {
                    throw new TypeError("Super expression must either be null or a function");
                  }
                  subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, writable: true, configurable: true } });
                  if (superClass) _setPrototypeOf(subClass, superClass);
                }
                function _setPrototypeOf(o, p) {
                  _setPrototypeOf = Object.setPrototypeOf || function _setPrototypeOf2(o2, p2) {
                    o2.__proto__ = p2;
                    return o2;
                  };
                  return _setPrototypeOf(o, p);
                }
                function _createSuper(Derived) {
                  var hasNativeReflectConstruct = _isNativeReflectConstruct();
                  return function _createSuperInternal() {
                    var Super = _getPrototypeOf(Derived), result;
                    if (hasNativeReflectConstruct) {
                      var NewTarget = _getPrototypeOf(this).constructor;
                      result = Reflect.construct(Super, arguments, NewTarget);
                    } else {
                      result = Super.apply(this, arguments);
                    }
                    return _possibleConstructorReturn(this, result);
                  };
                }
                function _possibleConstructorReturn(self, call) {
                  if (call && (clipboard_typeof(call) === "object" || typeof call === "function")) {
                    return call;
                  }
                  return _assertThisInitialized(self);
                }
                function _assertThisInitialized(self) {
                  if (self === void 0) {
                    throw new ReferenceError("this hasn't been initialised - super() hasn't been called");
                  }
                  return self;
                }
                function _isNativeReflectConstruct() {
                  if (typeof Reflect === "undefined" || !Reflect.construct) return false;
                  if (Reflect.construct.sham) return false;
                  if (typeof Proxy === "function") return true;
                  try {
                    Date.prototype.toString.call(Reflect.construct(Date, [], function() {
                    }));
                    return true;
                  } catch (e) {
                    return false;
                  }
                }
                function _getPrototypeOf(o) {
                  _getPrototypeOf = Object.setPrototypeOf ? Object.getPrototypeOf : function _getPrototypeOf2(o2) {
                    return o2.__proto__ || Object.getPrototypeOf(o2);
                  };
                  return _getPrototypeOf(o);
                }
                function getAttributeValue(suffix, element) {
                  var attribute = "data-clipboard-".concat(suffix);
                  if (!element.hasAttribute(attribute)) {
                    return;
                  }
                  return element.getAttribute(attribute);
                }
                var Clipboard = /* @__PURE__ */ (function(_Emitter) {
                  _inherits(Clipboard2, _Emitter);
                  var _super = _createSuper(Clipboard2);
                  function Clipboard2(trigger2, options) {
                    var _this;
                    _classCallCheck(this, Clipboard2);
                    _this = _super.call(this);
                    _this.resolveOptions(options);
                    _this.listenClick(trigger2);
                    return _this;
                  }
                  _createClass(Clipboard2, [{
                    key: "resolveOptions",
                    value: function resolveOptions() {
                      var options = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
                      this.action = typeof options.action === "function" ? options.action : this.defaultAction;
                      this.target = typeof options.target === "function" ? options.target : this.defaultTarget;
                      this.text = typeof options.text === "function" ? options.text : this.defaultText;
                      this.container = clipboard_typeof(options.container) === "object" ? options.container : document.body;
                    }
                    /**
                     * Adds a click event listener to the passed trigger.
                     * @param {String|HTMLElement|HTMLCollection|NodeList} trigger
                     */
                  }, {
                    key: "listenClick",
                    value: function listenClick(trigger2) {
                      var _this2 = this;
                      this.listener = listen_default()(trigger2, "click", function(e) {
                        return _this2.onClick(e);
                      });
                    }
                    /**
                     * Defines a new `ClipboardAction` on each click event.
                     * @param {Event} e
                     */
                  }, {
                    key: "onClick",
                    value: function onClick(e) {
                      var trigger2 = e.delegateTarget || e.currentTarget;
                      var action = this.action(trigger2) || "copy";
                      var text = actions_default({
                        action,
                        container: this.container,
                        target: this.target(trigger2),
                        text: this.text(trigger2)
                      });
                      this.emit(text ? "success" : "error", {
                        action,
                        text,
                        trigger: trigger2,
                        clearSelection: function clearSelection() {
                          if (trigger2) {
                            trigger2.focus();
                          }
                          window.getSelection().removeAllRanges();
                        }
                      });
                    }
                    /**
                     * Default `action` lookup function.
                     * @param {Element} trigger
                     */
                  }, {
                    key: "defaultAction",
                    value: function defaultAction(trigger2) {
                      return getAttributeValue("action", trigger2);
                    }
                    /**
                     * Default `target` lookup function.
                     * @param {Element} trigger
                     */
                  }, {
                    key: "defaultTarget",
                    value: function defaultTarget(trigger2) {
                      var selector = getAttributeValue("target", trigger2);
                      if (selector) {
                        return document.querySelector(selector);
                      }
                    }
                    /**
                     * Allow fire programmatically a copy action
                     * @param {String|HTMLElement} target
                     * @param {Object} options
                     * @returns Text copied.
                     */
                  }, {
                    key: "defaultText",
                    /**
                     * Default `text` lookup function.
                     * @param {Element} trigger
                     */
                    value: function defaultText(trigger2) {
                      return getAttributeValue("text", trigger2);
                    }
                    /**
                     * Destroy lifecycle.
                     */
                  }, {
                    key: "destroy",
                    value: function destroy() {
                      this.listener.destroy();
                    }
                  }], [{
                    key: "copy",
                    value: function copy(target) {
                      var options = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {
                        container: document.body
                      };
                      return actions_copy(target, options);
                    }
                    /**
                     * Allow fire programmatically a cut action
                     * @param {String|HTMLElement} target
                     * @returns Text cutted.
                     */
                  }, {
                    key: "cut",
                    value: function cut(target) {
                      return actions_cut(target);
                    }
                    /**
                     * Returns the support of the given action, or all actions if no action is
                     * given.
                     * @param {String} [action]
                     */
                  }, {
                    key: "isSupported",
                    value: function isSupported() {
                      var action = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : ["copy", "cut"];
                      var actions = typeof action === "string" ? [action] : action;
                      var support = !!document.queryCommandSupported;
                      actions.forEach(function(action2) {
                        support = support && !!document.queryCommandSupported(action2);
                      });
                      return support;
                    }
                  }]);
                  return Clipboard2;
                })(tiny_emitter_default());
                var clipboard2 = Clipboard;
              })
            ),
            /***/
            828: (
              /***/
              (function(module2) {
                var DOCUMENT_NODE_TYPE = 9;
                if (typeof Element !== "undefined" && !Element.prototype.matches) {
                  var proto = Element.prototype;
                  proto.matches = proto.matchesSelector || proto.mozMatchesSelector || proto.msMatchesSelector || proto.oMatchesSelector || proto.webkitMatchesSelector;
                }
                function closest(element, selector) {
                  while (element && element.nodeType !== DOCUMENT_NODE_TYPE) {
                    if (typeof element.matches === "function" && element.matches(selector)) {
                      return element;
                    }
                    element = element.parentNode;
                  }
                }
                module2.exports = closest;
              })
            ),
            /***/
            438: (
              /***/
              (function(module2, __unused_webpack_exports, __webpack_require__2) {
                var closest = __webpack_require__2(828);
                function _delegate(element, selector, type, callback, useCapture) {
                  var listenerFn = listener.apply(this, arguments);
                  element.addEventListener(type, listenerFn, useCapture);
                  return {
                    destroy: function() {
                      element.removeEventListener(type, listenerFn, useCapture);
                    }
                  };
                }
                function delegate(elements, selector, type, callback, useCapture) {
                  if (typeof elements.addEventListener === "function") {
                    return _delegate.apply(null, arguments);
                  }
                  if (typeof type === "function") {
                    return _delegate.bind(null, document).apply(null, arguments);
                  }
                  if (typeof elements === "string") {
                    elements = document.querySelectorAll(elements);
                  }
                  return Array.prototype.map.call(elements, function(element) {
                    return _delegate(element, selector, type, callback, useCapture);
                  });
                }
                function listener(element, selector, type, callback) {
                  return function(e) {
                    e.delegateTarget = closest(e.target, selector);
                    if (e.delegateTarget) {
                      callback.call(element, e);
                    }
                  };
                }
                module2.exports = delegate;
              })
            ),
            /***/
            879: (
              /***/
              (function(__unused_webpack_module, exports2) {
                exports2.node = function(value) {
                  return value !== void 0 && value instanceof HTMLElement && value.nodeType === 1;
                };
                exports2.nodeList = function(value) {
                  var type = Object.prototype.toString.call(value);
                  return value !== void 0 && (type === "[object NodeList]" || type === "[object HTMLCollection]") && "length" in value && (value.length === 0 || exports2.node(value[0]));
                };
                exports2.string = function(value) {
                  return typeof value === "string" || value instanceof String;
                };
                exports2.fn = function(value) {
                  var type = Object.prototype.toString.call(value);
                  return type === "[object Function]";
                };
              })
            ),
            /***/
            370: (
              /***/
              (function(module2, __unused_webpack_exports, __webpack_require__2) {
                var is = __webpack_require__2(879);
                var delegate = __webpack_require__2(438);
                function listen(target, type, callback) {
                  if (!target && !type && !callback) {
                    throw new Error("Missing required arguments");
                  }
                  if (!is.string(type)) {
                    throw new TypeError("Second argument must be a String");
                  }
                  if (!is.fn(callback)) {
                    throw new TypeError("Third argument must be a Function");
                  }
                  if (is.node(target)) {
                    return listenNode(target, type, callback);
                  } else if (is.nodeList(target)) {
                    return listenNodeList(target, type, callback);
                  } else if (is.string(target)) {
                    return listenSelector(target, type, callback);
                  } else {
                    throw new TypeError("First argument must be a String, HTMLElement, HTMLCollection, or NodeList");
                  }
                }
                function listenNode(node, type, callback) {
                  node.addEventListener(type, callback);
                  return {
                    destroy: function() {
                      node.removeEventListener(type, callback);
                    }
                  };
                }
                function listenNodeList(nodeList, type, callback) {
                  Array.prototype.forEach.call(nodeList, function(node) {
                    node.addEventListener(type, callback);
                  });
                  return {
                    destroy: function() {
                      Array.prototype.forEach.call(nodeList, function(node) {
                        node.removeEventListener(type, callback);
                      });
                    }
                  };
                }
                function listenSelector(selector, type, callback) {
                  return delegate(document.body, selector, type, callback);
                }
                module2.exports = listen;
              })
            ),
            /***/
            817: (
              /***/
              (function(module2) {
                function select(element) {
                  var selectedText;
                  if (element.nodeName === "SELECT") {
                    element.focus();
                    selectedText = element.value;
                  } else if (element.nodeName === "INPUT" || element.nodeName === "TEXTAREA") {
                    var isReadOnly = element.hasAttribute("readonly");
                    if (!isReadOnly) {
                      element.setAttribute("readonly", "");
                    }
                    element.select();
                    element.setSelectionRange(0, element.value.length);
                    if (!isReadOnly) {
                      element.removeAttribute("readonly");
                    }
                    selectedText = element.value;
                  } else {
                    if (element.hasAttribute("contenteditable")) {
                      element.focus();
                    }
                    var selection = window.getSelection();
                    var range = document.createRange();
                    range.selectNodeContents(element);
                    selection.removeAllRanges();
                    selection.addRange(range);
                    selectedText = selection.toString();
                  }
                  return selectedText;
                }
                module2.exports = select;
              })
            ),
            /***/
            279: (
              /***/
              (function(module2) {
                function E() {
                }
                E.prototype = {
                  on: function(name, callback, ctx) {
                    var e = this.e || (this.e = {});
                    (e[name] || (e[name] = [])).push({
                      fn: callback,
                      ctx
                    });
                    return this;
                  },
                  once: function(name, callback, ctx) {
                    var self = this;
                    function listener() {
                      self.off(name, listener);
                      callback.apply(ctx, arguments);
                    }
                    ;
                    listener._ = callback;
                    return this.on(name, listener, ctx);
                  },
                  emit: function(name) {
                    var data = [].slice.call(arguments, 1);
                    var evtArr = ((this.e || (this.e = {}))[name] || []).slice();
                    var i = 0;
                    var len = evtArr.length;
                    for (i; i < len; i++) {
                      evtArr[i].fn.apply(evtArr[i].ctx, data);
                    }
                    return this;
                  },
                  off: function(name, callback) {
                    var e = this.e || (this.e = {});
                    var evts = e[name];
                    var liveEvents = [];
                    if (evts && callback) {
                      for (var i = 0, len = evts.length; i < len; i++) {
                        if (evts[i].fn !== callback && evts[i].fn._ !== callback)
                          liveEvents.push(evts[i]);
                      }
                    }
                    liveEvents.length ? e[name] = liveEvents : delete e[name];
                    return this;
                  }
                };
                module2.exports = E;
                module2.exports.TinyEmitter = E;
              })
            )
            /******/
          };
          var __webpack_module_cache__ = {};
          function __webpack_require__(moduleId) {
            if (__webpack_module_cache__[moduleId]) {
              return __webpack_module_cache__[moduleId].exports;
            }
            var module2 = __webpack_module_cache__[moduleId] = {
              /******/
              // no module.id needed
              /******/
              // no module.loaded needed
              /******/
              exports: {}
              /******/
            };
            __webpack_modules__[moduleId](module2, module2.exports, __webpack_require__);
            return module2.exports;
          }
          !(function() {
            __webpack_require__.n = function(module2) {
              var getter = module2 && module2.__esModule ? (
                /******/
                function() {
                  return module2["default"];
                }
              ) : (
                /******/
                function() {
                  return module2;
                }
              );
              __webpack_require__.d(getter, { a: getter });
              return getter;
            };
          })();
          !(function() {
            __webpack_require__.d = function(exports2, definition) {
              for (var key in definition) {
                if (__webpack_require__.o(definition, key) && !__webpack_require__.o(exports2, key)) {
                  Object.defineProperty(exports2, key, { enumerable: true, get: definition[key] });
                }
              }
            };
          })();
          !(function() {
            __webpack_require__.o = function(obj, prop) {
              return Object.prototype.hasOwnProperty.call(obj, prop);
            };
          })();
          return __webpack_require__(686);
        })().default
      );
    });
  }
});

// node_modules/bootstrap/js/dist/dom/data.js
var elementMap = /* @__PURE__ */ new Map();
var data_default = {
  set(element, key, instance) {
    if (!elementMap.has(element)) elementMap.set(element, /* @__PURE__ */ new Map());
    const instanceMap = elementMap.get(element);
    if (!instanceMap.has(key) && instanceMap.size !== 0) {
      console.error(`Bootstrap doesn't allow more than one instance per element. Bound instance: ${[...instanceMap.keys()][0]}.`);
      return;
    }
    instanceMap.set(key, instance);
  },
  get(element, key) {
    if (element && elementMap.has(element)) return elementMap.get(element).get(key) || null;
    return null;
  },
  getAny(element) {
    if (element && elementMap.has(element)) return elementMap.get(element).values().next().value || null;
    return null;
  },
  remove(element, key) {
    if (!elementMap.has(element)) return;
    const instanceMap = elementMap.get(element);
    instanceMap.delete(key);
    if (instanceMap.size === 0) elementMap.delete(element);
  }
};

// node_modules/bootstrap/js/dist/dom/event-handler.js
var namespaceRegex = /[^.]*(?=\..*)\.|.*/;
var stripNameRegex = /\..*/;
var stripUidRegex = /::\d+$/;
var eventRegistry = {};
var uidEvent = 1;
var customEvents = {
  mouseenter: "mouseover",
  mouseleave: "mouseout"
};
var nativeEvents = /* @__PURE__ */ new Set([
  "click",
  "dblclick",
  "mouseup",
  "mousedown",
  "contextmenu",
  "mousewheel",
  "DOMMouseScroll",
  "mouseover",
  "mouseout",
  "mousemove",
  "selectstart",
  "selectend",
  "keydown",
  "keypress",
  "keyup",
  "input",
  "paste",
  "orientationchange",
  "touchstart",
  "touchmove",
  "touchend",
  "touchcancel",
  "pointerdown",
  "pointermove",
  "pointerup",
  "pointerleave",
  "pointercancel",
  "gesturestart",
  "gesturechange",
  "gestureend",
  "focus",
  "blur",
  "change",
  "reset",
  "select",
  "submit",
  "focusin",
  "focusout",
  "load",
  "unload",
  "beforeunload",
  "resize",
  "move",
  "DOMContentLoaded",
  "readystatechange",
  "error",
  "abort",
  "scroll",
  "scrollend"
]);
function makeEventUid(element, uid) {
  return uid && `${uid}::${uidEvent++}` || element.uidEvent || uidEvent++;
}
function getElementEvents(element) {
  const uid = makeEventUid(element);
  element.uidEvent = uid;
  eventRegistry[uid] = eventRegistry[uid] || {};
  return eventRegistry[uid];
}
function isMouseEventWithinTarget(event) {
  const { delegateTarget, relatedTarget } = event;
  return Boolean(relatedTarget && delegateTarget.contains(relatedTarget));
}
function bootstrapHandler(element, fn, handlerTypeEvent) {
  const isCustomMouseEvent = handlerTypeEvent in customEvents;
  return function handler(event) {
    const bootstrapEvent = hydrateObj(event, { delegateTarget: element });
    if (isCustomMouseEvent && isMouseEventWithinTarget(bootstrapEvent)) return;
    if (handler.oneOff) EventHandler.off(element, handlerTypeEvent, fn);
    return fn.apply(element, [bootstrapEvent]);
  };
}
function bootstrapDelegationHandler(element, selector, fn, handlerTypeEvent) {
  const isCustomMouseEvent = handlerTypeEvent in customEvents;
  return function handler(event) {
    const domElements = element.querySelectorAll(selector);
    for (let { target } = event; target && target !== this; target = target.parentNode) for (const domElement of domElements) {
      if (domElement !== target) continue;
      const bootstrapEvent = hydrateObj(event, { delegateTarget: target });
      if (isCustomMouseEvent && isMouseEventWithinTarget(bootstrapEvent)) return;
      if (handler.oneOff) EventHandler.off(element, handlerTypeEvent, selector, fn);
      return fn.apply(target, [bootstrapEvent]);
    }
  };
}
function findHandler(events, callable, handlerTypeEvent, delegationSelector = null) {
  return Object.values(events).find((event) => event.callable === callable && event.handlerTypeEvent === handlerTypeEvent && event.delegationSelector === delegationSelector);
}
function normalizeParameters(originalTypeEvent, handler, delegationFunction) {
  const isDelegated = typeof handler === "string";
  const callable = isDelegated ? delegationFunction : handler || delegationFunction;
  const baseTypeEvent = originalTypeEvent.replace(stripNameRegex, "");
  let typeEvent = customEvents[baseTypeEvent] || baseTypeEvent;
  if (!nativeEvents.has(typeEvent)) typeEvent = originalTypeEvent;
  const handlerTypeEvent = baseTypeEvent in customEvents ? baseTypeEvent : typeEvent;
  return {
    isDelegated,
    callable,
    typeEvent,
    handlerTypeEvent
  };
}
function addHandler(element, originalTypeEvent, handler, delegationFunction, oneOff) {
  if (typeof originalTypeEvent !== "string" || !element) return;
  const { isDelegated, callable, typeEvent, handlerTypeEvent } = normalizeParameters(originalTypeEvent, handler, delegationFunction);
  const events = getElementEvents(element);
  const handlers = events[typeEvent] || (events[typeEvent] = {});
  const previousFunction = findHandler(handlers, callable, handlerTypeEvent, isDelegated ? handler : null);
  if (previousFunction) {
    previousFunction.oneOff = previousFunction.oneOff && oneOff;
    return;
  }
  const uid = makeEventUid(callable, originalTypeEvent.replace(namespaceRegex, ""));
  const fn = isDelegated ? bootstrapDelegationHandler(element, handler, callable, handlerTypeEvent) : bootstrapHandler(element, callable, handlerTypeEvent);
  fn.delegationSelector = isDelegated ? handler : null;
  fn.callable = callable;
  fn.handlerTypeEvent = handlerTypeEvent;
  fn.oneOff = oneOff;
  fn.uidEvent = uid;
  handlers[uid] = fn;
  element.addEventListener(typeEvent, fn, isDelegated);
}
function removeHandler(element, events, typeEvent, handler) {
  element.removeEventListener(typeEvent, handler, Boolean(handler.delegationSelector));
  delete events[typeEvent][handler.uidEvent];
}
function removeNamespacedHandlers(element, events, typeEvent, namespace) {
  const storeElementEvent = events[typeEvent] || {};
  for (const [handlerKey, event] of Object.entries(storeElementEvent)) if (handlerKey.includes(namespace)) removeHandler(element, events, typeEvent, event);
}
function trigger(element, event, args) {
  if (typeof event !== "string" || !element) return null;
  const evt = hydrateObj(new Event(event, {
    bubbles: true,
    cancelable: true
  }), args);
  element.dispatchEvent(evt);
  return evt;
}
var EventHandler = {
  on(element, event, handler, delegationFunction) {
    addHandler(element, event, handler, delegationFunction, false);
  },
  one(element, event, handler, delegationFunction) {
    addHandler(element, event, handler, delegationFunction, true);
  },
  off(element, originalTypeEvent, handler, delegationFunction) {
    if (typeof originalTypeEvent !== "string" || !element) return;
    const { isDelegated, callable, typeEvent, handlerTypeEvent } = normalizeParameters(originalTypeEvent, handler, delegationFunction);
    const inNamespace = typeEvent !== originalTypeEvent && handlerTypeEvent !== originalTypeEvent;
    const events = getElementEvents(element);
    const storeElementEvent = events[typeEvent] || {};
    const isNamespace = originalTypeEvent.startsWith(".");
    if (typeof callable !== "undefined") {
      if (!Object.keys(storeElementEvent).length) return;
      const fn = findHandler(storeElementEvent, callable, handlerTypeEvent, isDelegated ? handler : null);
      if (fn) removeHandler(element, events, typeEvent, fn);
      return;
    }
    if (isNamespace) for (const elementEvent of Object.keys(events)) removeNamespacedHandlers(element, events, elementEvent, originalTypeEvent.slice(1));
    for (const [keyHandlers, event] of Object.entries(storeElementEvent)) {
      const handlerKey = keyHandlers.replace(stripUidRegex, "");
      if (event.handlerTypeEvent === handlerTypeEvent && (!inNamespace || originalTypeEvent.includes(handlerKey))) removeHandler(element, events, typeEvent, event);
    }
  },
  trigger
};
function hydrateObj(obj, meta = {}) {
  for (const [key, value] of Object.entries(meta)) try {
    obj[key] = value;
  } catch {
    Object.defineProperty(obj, key, {
      configurable: true,
      get() {
        return value;
      }
    });
  }
  return obj;
}

// node_modules/bootstrap/js/dist/dom/manipulator.js
function normalizeData(value) {
  if (value === "true") return true;
  if (value === "false") return false;
  if (value === Number(value).toString()) return Number(value);
  if (value === "" || value === "null") return null;
  if (typeof value !== "string") return value;
  try {
    return JSON.parse(decodeURIComponent(value));
  } catch {
    return value;
  }
}
function normalizeDataKey(key) {
  return key.replace(/[A-Z]/g, (chr) => `-${chr.toLowerCase()}`);
}
var Manipulator = {
  setDataAttribute(element, key, value) {
    element.setAttribute(`data-bs-${normalizeDataKey(key)}`, value);
  },
  removeDataAttribute(element, key) {
    element.removeAttribute(`data-bs-${normalizeDataKey(key)}`);
  },
  getDataAttributes(element) {
    if (!element) return {};
    const attributes = {};
    const bsKeys = Object.keys(element.dataset).filter((key) => key.startsWith("bs") && !key.startsWith("bsConfig"));
    for (const key of bsKeys) {
      let pureKey = key.replace(/^bs/, "");
      pureKey = pureKey.charAt(0).toLowerCase() + pureKey.slice(1);
      attributes[pureKey] = normalizeData(element.dataset[key]);
    }
    return attributes;
  },
  getDataAttribute(element, key) {
    return normalizeData(element.getAttribute(`data-bs-${normalizeDataKey(key)}`));
  }
};

// node_modules/bootstrap/js/dist/util/index.js
var MAX_UID = 1e6;
var MILLISECONDS_MULTIPLIER = 1e3;
var TRANSITION_END = "transitionend";
var parseSelector = (selector) => {
  if (selector && window.CSS && window.CSS.escape) selector = selector.replace(/#([^\s"#']+)/g, (match, id) => `#${CSS.escape(id)}`);
  return selector;
};
var toType = (object) => {
  if (object === null || object === void 0) return `${object}`;
  return Object.prototype.toString.call(object).match(/\s([a-z]+)/i)[1].toLowerCase();
};
var getUID = (prefix) => {
  do
    prefix += Math.floor(Math.random() * MAX_UID);
  while (document.getElementById(prefix));
  return prefix;
};
var getTransitionDurationFromElement = (element) => {
  if (!element) return 0;
  let { transitionDuration, transitionDelay } = window.getComputedStyle(element);
  if (!Number.parseFloat(transitionDuration) && !Number.parseFloat(transitionDelay)) return 0;
  transitionDuration = transitionDuration.split(",")[0];
  transitionDelay = transitionDelay.split(",")[0];
  return (Number.parseFloat(transitionDuration) + Number.parseFloat(transitionDelay)) * MILLISECONDS_MULTIPLIER;
};
var triggerTransitionEnd = (element) => {
  element.dispatchEvent(new Event(TRANSITION_END));
};
var isElement = (object) => {
  if (!object || typeof object !== "object") return false;
  return typeof object.nodeType !== "undefined";
};
var getElement = (object) => {
  if (isElement(object)) return object;
  if (typeof object === "string" && object.length > 0) return document.querySelector(parseSelector(object));
  return null;
};
var isVisible = (element) => {
  if (!isElement(element) || element.getClientRects().length === 0) return false;
  const elementIsVisible = getComputedStyle(element).getPropertyValue("visibility") === "visible";
  const closedDetails = element.closest("details:not([open])");
  if (!closedDetails) return elementIsVisible;
  if (closedDetails !== element) {
    const summary = element.closest("summary");
    if (summary && summary.parentNode !== closedDetails) return false;
    if (summary === null) return false;
  }
  return elementIsVisible;
};
var isDisabled = (element) => {
  if (!element || element.nodeType !== Node.ELEMENT_NODE) return true;
  if (element.classList.contains("disabled")) return true;
  const disableableElement = element;
  if (typeof disableableElement.disabled !== "undefined") return disableableElement.disabled;
  return element.hasAttribute("disabled") && element.getAttribute("disabled") !== "false";
};
var setAriaAttribute = (element, name, value) => {
  element.setAttribute(name, String(value));
};
var findShadowRoot = (element) => {
  if (!document.documentElement.attachShadow) return null;
  if (typeof element.getRootNode === "function") {
    const root = element.getRootNode();
    return root instanceof ShadowRoot ? root : null;
  }
  if (element instanceof ShadowRoot) return element;
  if (!element.parentNode) return null;
  return findShadowRoot(element.parentNode);
};
var noop = () => {
};
var isRTL = () => document.documentElement.dir === "rtl";
var execute = (possibleCallback, args = [], defaultValue = possibleCallback) => {
  return typeof possibleCallback === "function" ? possibleCallback.call(...args) : defaultValue;
};
var executeAfterTransition = (callback, transitionElement, waitForTransition = true) => {
  if (!waitForTransition) {
    execute(callback);
    return;
  }
  const emulatedDuration = getTransitionDurationFromElement(transitionElement) + 5;
  let called = false;
  const handler = ({ target }) => {
    if (target !== transitionElement) return;
    called = true;
    transitionElement.removeEventListener(TRANSITION_END, handler);
    execute(callback);
  };
  transitionElement.addEventListener(TRANSITION_END, handler);
  setTimeout(() => {
    if (!called) triggerTransitionEnd(transitionElement);
  }, emulatedDuration);
};
var getNextActiveElement = (list, activeElement, shouldGetNext, isCycleAllowed) => {
  const listLength = list.length;
  let index = list.indexOf(activeElement);
  if (index === -1) return !shouldGetNext && isCycleAllowed ? list[listLength - 1] : list[0];
  index += shouldGetNext ? 1 : -1;
  if (isCycleAllowed) index = (index + listLength) % listLength;
  return list[Math.max(0, Math.min(index, listLength - 1))];
};

// node_modules/bootstrap/js/dist/util/config.js
var Config = class {
  static get Default() {
    return {};
  }
  static get DefaultType() {
    return {};
  }
  static get NAME() {
    throw new Error('You have to implement the static method "NAME", for each component!');
  }
  _getConfig(config) {
    config = this._mergeConfigObj(config);
    config = this._configAfterMerge(config);
    this._typeCheckConfig(config);
    return config;
  }
  _configAfterMerge(config) {
    return config;
  }
  _mergeConfigObj(config, element) {
    const jsonConfig = isElement(element) ? Manipulator.getDataAttribute(element, "config") : {};
    return {
      ...this.constructor.Default,
      ...typeof jsonConfig === "object" ? jsonConfig : {},
      ...isElement(element) ? Manipulator.getDataAttributes(element) : {},
      ...typeof config === "object" ? config : {}
    };
  }
  _typeCheckConfig(config, configTypes = this.constructor.DefaultType) {
    for (const [property, expectedTypes] of Object.entries(configTypes)) {
      const value = config[property];
      const valueType = isElement(value) ? "element" : toType(value);
      if (!new RegExp(expectedTypes).test(valueType)) throw new TypeError(`${this.constructor.NAME.toUpperCase()}: Option "${property}" provided type "${valueType}" but expected type "${expectedTypes}".`);
    }
  }
};

// node_modules/bootstrap/js/dist/base-component.js
var VERSION = "6.0.0-alpha.1";
var BaseComponent = class extends Config {
  constructor(element, config) {
    super();
    element = getElement(element);
    if (!element) return;
    this._element = element;
    this._config = this._getConfig(config);
    const existingInstance = data_default.get(this._element, this.constructor.DATA_KEY);
    if (existingInstance) existingInstance.dispose();
    data_default.set(this._element, this.constructor.DATA_KEY, this);
  }
  dispose() {
    data_default.remove(this._element, this.constructor.DATA_KEY);
    EventHandler.off(this._element, this.constructor.EVENT_KEY);
    for (const propertyName of Object.getOwnPropertyNames(this)) this[propertyName] = null;
  }
  _queueCallback(callback, element, isAnimated = true) {
    return new Promise((resolve) => {
      executeAfterTransition(() => {
        if (this._element) callback();
        resolve();
      }, element, isAnimated);
    });
  }
  _getConfig(config) {
    config = this._mergeConfigObj(config, this._element);
    config = this._configAfterMerge(config);
    this._typeCheckConfig(config);
    return config;
  }
  static getInstance(element) {
    return data_default.get(getElement(element), this.DATA_KEY);
  }
  static getOrCreateInstance(element, config = {}) {
    return this.getInstance(element) || new this(element, typeof config === "object" ? config : null);
  }
  static get VERSION() {
    return VERSION;
  }
  static get DATA_KEY() {
    return `bs.${this.NAME}`;
  }
  static get EVENT_KEY() {
    return `.${this.DATA_KEY}`;
  }
  static eventName(name) {
    return `${name}${this.EVENT_KEY}`;
  }
};

// node_modules/bootstrap/js/dist/dom/selector-engine.js
var getSelector = (element) => {
  let selector = element.getAttribute("data-bs-target");
  if (!selector || selector === "#") {
    let hrefAttribute = element.getAttribute("href");
    if (!hrefAttribute || !hrefAttribute.includes("#") && !hrefAttribute.startsWith(".")) return null;
    if (hrefAttribute.includes("#") && !hrefAttribute.startsWith("#")) hrefAttribute = `#${hrefAttribute.split("#")[1]}`;
    selector = hrefAttribute && hrefAttribute !== "#" ? hrefAttribute.trim() : null;
  }
  return selector ? selector.split(",").map((sel) => parseSelector(sel)).join(",") : null;
};
var SelectorEngine = {
  find(selector, element = document.documentElement) {
    return [...Element.prototype.querySelectorAll.call(element, selector)];
  },
  findOne(selector, element = document.documentElement) {
    return Element.prototype.querySelector.call(element, selector);
  },
  parents(element, selector) {
    const parents = [];
    let ancestor = element.parentNode.closest(selector);
    while (ancestor) {
      parents.push(ancestor);
      ancestor = ancestor.parentNode.closest(selector);
    }
    return parents;
  },
  closest(element, selector) {
    return Element.prototype.closest.call(element, selector);
  },
  prev(element, selector) {
    let previous = element.previousElementSibling;
    while (previous) {
      if (previous.matches(selector)) return [previous];
      previous = previous.previousElementSibling;
    }
    return [];
  },
  next(element, selector) {
    let next = element.nextElementSibling;
    while (next) {
      if (next.matches(selector)) return [next];
      next = next.nextElementSibling;
    }
    return [];
  },
  getSelectorFromElement(element) {
    const selector = getSelector(element);
    if (selector) return SelectorEngine.findOne(selector) ? selector : null;
    return null;
  },
  getElementFromSelector(element) {
    const selector = getSelector(element);
    return selector ? SelectorEngine.findOne(selector) : null;
  },
  getMultipleElementsFromSelector(element) {
    const selector = getSelector(element);
    return selector ? SelectorEngine.find(selector) : [];
  }
};

// node_modules/bootstrap/js/dist/dialog-base.js
var CLASS_NAME_OPEN = "dialog-open";
var DialogBase = class extends BaseComponent {
  constructor(element, config) {
    super(element, config);
    this._isTransitioning = false;
    this._openedAsModal = false;
    this._addDialogListeners();
  }
  static get NAME() {
    return "dialogbase";
  }
  toggle(relatedTarget) {
    return this._element.open ? this.hide() : this.show(relatedTarget);
  }
  async show(relatedTarget) {
    if (!this._canShow(relatedTarget)) return;
    await this._show(relatedTarget);
  }
  async hide() {
    if (!this._canHide()) return;
    await this._hide();
  }
  _canShow(relatedTarget) {
    if (this._element.open || this._isTransitioning) return false;
    return !EventHandler.trigger(this._element, this.constructor.eventName("show"), { relatedTarget }).defaultPrevented;
  }
  async _show(relatedTarget) {
    this._isTransitioning = true;
    this._onBeforeShow();
    const { modal, preventBodyScroll } = this._getShowOptions();
    this._showElement({
      modal,
      preventBodyScroll
    });
    await this._queueCallback(() => {
      this._isTransitioning = false;
      EventHandler.trigger(this._element, this.constructor.eventName("shown"), { relatedTarget });
    }, this._element, this._isAnimated());
  }
  _canHide() {
    if (!this._element.open || this._isTransitioning) return false;
    return !EventHandler.trigger(this._element, this.constructor.eventName("hide")).defaultPrevented;
  }
  async _hide() {
    this._isTransitioning = true;
    this._hideElement();
    await this._queueCallback(() => {
      if (this._element.open) this._closeAndCleanup();
      this._element.classList.remove("hiding");
      this._onAfterHide();
      this._isTransitioning = false;
      EventHandler.trigger(this._element, this.constructor.eventName("hidden"));
    }, this._element, this._isAnimated());
  }
  dispose() {
    if (this._element.open) this._closeAndCleanup();
    EventHandler.off(this._element, "cancel", this._cancelHandler);
    super.dispose();
  }
  _getShowOptions() {
    return {
      modal: true,
      preventBodyScroll: true
    };
  }
  _onBeforeShow() {
  }
  _onAfterHide() {
  }
  _isAnimated() {
    return !this._element.classList.contains(this._getInstantClassName());
  }
  _getInstantClassName() {
    return "dialog-instant";
  }
  _getStaticClassName() {
    return "dialog-static";
  }
  _onCancel() {
  }
  _showElement({ modal = true, preventBodyScroll = true } = {}) {
    this._openedAsModal = modal;
    if (modal) this._element.showModal();
    else this._element.show();
    if (preventBodyScroll) document.documentElement.classList.add(CLASS_NAME_OPEN);
  }
  _hideElement() {
    this._hideChildComponents();
    this._element.classList.add("hiding");
    if (!this._shouldDeferClose()) this._closeAndCleanup();
  }
  _closeAndCleanup() {
    this._element.close();
    this._openedAsModal = false;
    if (!document.querySelector("dialog[open]:modal")) document.documentElement.classList.remove(CLASS_NAME_OPEN);
  }
  _shouldDeferClose() {
    return false;
  }
  _triggerBackdropTransition() {
    if (EventHandler.trigger(this._element, this.constructor.eventName("hidePrevented")).defaultPrevented) return;
    const staticClass = this._getStaticClassName();
    this._element.classList.add(staticClass);
    this._queueCallback(() => {
      this._element.classList.remove(staticClass);
    }, this._element);
  }
  _hideChildComponents() {
    for (const el of SelectorEngine.find('[data-bs-toggle="tooltip"], [data-bs-toggle="popover"]', this._element)) {
      const instance = data_default.getAny(el);
      if (instance && typeof instance.hide === "function") instance.hide();
    }
    for (const el of SelectorEngine.find(".toast.show", this._element)) {
      const instance = data_default.getAny(el);
      if (instance && typeof instance.hide === "function") instance.hide();
    }
  }
  _addDialogListeners() {
    const eventKey = this.constructor.EVENT_KEY;
    this._cancelHandler = (event) => {
      event.preventDefault();
      if (!this._config.keyboard) {
        this._triggerBackdropTransition();
        return;
      }
      this._onCancel();
      this.hide();
    };
    EventHandler.on(this._element, "cancel", this._cancelHandler);
    EventHandler.on(this._element, `keydown${eventKey}`, (event) => {
      if (event.key !== "Escape" || this._openedAsModal) return;
      event.preventDefault();
      if (!this._config.keyboard) return;
      this._onCancel();
      this.hide();
    });
    EventHandler.on(this._element, `click${eventKey}`, (event) => {
      if (event.target !== this._element || !this._openedAsModal) return;
      if (this._config.backdrop === false) return;
      if (this._config.backdrop === "static") {
        this._triggerBackdropTransition();
        return;
      }
      this.hide();
    });
  }
};

// node_modules/bootstrap/js/dist/util/component-functions.js
var enableDismissTrigger = (component, method = "hide") => {
  const clickEvent = `click.dismiss${component.EVENT_KEY}`;
  const name = component.NAME;
  EventHandler.on(document, clickEvent, `[data-bs-dismiss="${name}"]`, function(event) {
    if (["A", "AREA"].includes(this.tagName)) event.preventDefault();
    if (isDisabled(this)) return;
    const target = SelectorEngine.getElementFromSelector(this) || this.closest(`.${name}, [class*=":${name}"]`);
    component.getOrCreateInstance(target)[method]();
  });
};

// node_modules/bootstrap/js/dist/dialog.js
var NAME = "dialog";
var EVENT_KEY = `.bs.dialog`;
var DATA_API_KEY = ".data-api";
var EVENT_SHOW = `show${EVENT_KEY}`;
var EVENT_HIDDEN = `hidden${EVENT_KEY}`;
var EVENT_CANCEL = `cancel${EVENT_KEY}`;
var EVENT_CLICK_DATA_API = `click${EVENT_KEY}${DATA_API_KEY}`;
var CLASS_NAME_NONMODAL = "dialog-nonmodal";
var CLASS_NAME_INSTANT = "dialog-instant";
var CLASS_NAME_SWAP_IN = "dialog-swap-in";
var SELECTOR_DATA_TOGGLE = '[data-bs-toggle="dialog"]';
var Default = {
  backdrop: true,
  keyboard: true,
  modal: true
};
var DefaultType = {
  backdrop: "(boolean|string)",
  keyboard: "boolean",
  modal: "boolean"
};
var Dialog = class Dialog2 extends DialogBase {
  constructor(element, config) {
    super(element, config);
  }
  static get Default() {
    return Default;
  }
  static get DefaultType() {
    return DefaultType;
  }
  static get NAME() {
    return NAME;
  }
  static {
    EventHandler.on(document, EVENT_CLICK_DATA_API, SELECTOR_DATA_TOGGLE, function(event) {
      Dialog2._handleDataApiClick(this, event);
    });
  }
  static _handleDataApiClick(trigger2, event) {
    const target = SelectorEngine.getElementFromSelector(trigger2);
    if (!target) return;
    if (["A", "AREA"].includes(trigger2.tagName)) event.preventDefault();
    const config = Manipulator.getDataAttributes(trigger2);
    const currentDialog = trigger2.closest("dialog[open]");
    if (!(currentDialog && currentDialog !== target)) {
      EventHandler.one(target, EVENT_SHOW, (showEvent) => {
        if (!showEvent.defaultPrevented) Dialog2._registerFocusRestoration(target, trigger2);
      });
      Dialog2.getOrCreateInstance(target, config).toggle(trigger2);
      return;
    }
    if (currentDialog.classList.contains(CLASS_NAME_SWAP_IN)) return;
    const newDialog = Dialog2.getOrCreateInstance(target, config);
    const currentInstance = Dialog2.getOrCreateInstance(currentDialog);
    if (!newDialog._canShow(trigger2) || !currentInstance._canHide()) return;
    Dialog2._registerFocusRestoration(target, trigger2);
    target.classList.add(CLASS_NAME_SWAP_IN);
    EventHandler.one(target, `shown${EVENT_KEY}`, () => {
      target.classList.remove(CLASS_NAME_SWAP_IN);
    });
    newDialog._show(trigger2);
    currentDialog.classList.add(CLASS_NAME_INSTANT);
    EventHandler.one(currentDialog, EVENT_HIDDEN, () => {
      currentDialog.classList.remove(CLASS_NAME_INSTANT);
    });
    currentInstance._hide();
  }
  static _registerFocusRestoration(target, trigger2) {
    EventHandler.one(target, EVENT_HIDDEN, () => {
      if (isVisible(trigger2)) trigger2.focus({ preventScroll: true });
    });
  }
  handleUpdate() {
  }
  _getShowOptions() {
    return {
      modal: this._config.modal,
      preventBodyScroll: this._config.modal
    };
  }
  _onBeforeShow() {
    if (!this._config.modal) this._element.classList.add(CLASS_NAME_NONMODAL);
  }
  _onAfterHide() {
    this._element.classList.remove(CLASS_NAME_NONMODAL);
  }
  _shouldDeferClose() {
    return this._isAnimated();
  }
  _onCancel() {
    EventHandler.trigger(this._element, EVENT_CANCEL);
  }
};
enableDismissTrigger(Dialog);

// node_modules/bootstrap/js/dist/util/swipe.js
var NAME2 = "swipe";
var EVENT_KEY2 = ".bs.swipe";
var EVENT_TOUCHSTART = `touchstart${EVENT_KEY2}`;
var EVENT_TOUCHMOVE = `touchmove${EVENT_KEY2}`;
var EVENT_TOUCHEND = `touchend${EVENT_KEY2}`;
var EVENT_POINTERDOWN = `pointerdown${EVENT_KEY2}`;
var EVENT_POINTERUP = `pointerup${EVENT_KEY2}`;
var POINTER_TYPE_TOUCH = "touch";
var POINTER_TYPE_PEN = "pen";
var CLASS_NAME_POINTER_EVENT = "pointer-event";
var SWIPE_THRESHOLD = 40;
var Default2 = {
  endCallback: null,
  leftCallback: null,
  rightCallback: null,
  upCallback: null,
  downCallback: null
};
var DefaultType2 = {
  endCallback: "(function|null)",
  leftCallback: "(function|null)",
  rightCallback: "(function|null)",
  upCallback: "(function|null)",
  downCallback: "(function|null)"
};
var Swipe = class Swipe2 extends Config {
  constructor(element, config) {
    super();
    this._element = element;
    if (!element || !Swipe2.isSupported()) return;
    this._config = this._getConfig(config);
    this._deltaX = 0;
    this._deltaY = 0;
    this._supportPointerEvents = Boolean(window.PointerEvent);
    this._initEvents();
  }
  static get Default() {
    return Default2;
  }
  static get DefaultType() {
    return DefaultType2;
  }
  static get NAME() {
    return NAME2;
  }
  dispose() {
    EventHandler.off(this._element, EVENT_KEY2);
  }
  _start(event) {
    if (!this._supportPointerEvents) {
      this._deltaX = event.touches[0].clientX;
      this._deltaY = event.touches[0].clientY;
      return;
    }
    if (this._eventIsPointerPenTouch(event)) {
      this._deltaX = event.clientX;
      this._deltaY = event.clientY;
    }
  }
  _end(event) {
    if (this._eventIsPointerPenTouch(event)) {
      this._deltaX = event.clientX - this._deltaX;
      this._deltaY = event.clientY - this._deltaY;
    }
    this._handleSwipe();
    execute(this._config.endCallback);
  }
  _move(event) {
    if (event.touches && event.touches.length > 1) {
      this._deltaX = 0;
      this._deltaY = 0;
      return;
    }
    this._deltaX = event.touches[0].clientX - this._deltaX;
    this._deltaY = event.touches[0].clientY - this._deltaY;
  }
  _handleSwipe() {
    const absDeltaX = Math.abs(this._deltaX);
    const absDeltaY = Math.abs(this._deltaY);
    if (absDeltaY > absDeltaX && absDeltaY > SWIPE_THRESHOLD) {
      const direction = this._deltaY > 0 ? "down" : "up";
      this._deltaX = 0;
      this._deltaY = 0;
      execute(direction === "down" ? this._config.downCallback : this._config.upCallback);
      return;
    }
    if (absDeltaX > SWIPE_THRESHOLD) {
      const direction = absDeltaX / this._deltaX;
      this._deltaX = 0;
      this._deltaY = 0;
      if (!direction) return;
      execute(direction > 0 ? this._config.rightCallback : this._config.leftCallback);
      return;
    }
    this._deltaX = 0;
    this._deltaY = 0;
  }
  _initEvents() {
    if (this._supportPointerEvents) {
      EventHandler.on(this._element, EVENT_POINTERDOWN, (event) => this._start(event));
      EventHandler.on(this._element, EVENT_POINTERUP, (event) => this._end(event));
      this._element.classList.add(CLASS_NAME_POINTER_EVENT);
    } else {
      EventHandler.on(this._element, EVENT_TOUCHSTART, (event) => this._start(event));
      EventHandler.on(this._element, EVENT_TOUCHMOVE, (event) => this._move(event));
      EventHandler.on(this._element, EVENT_TOUCHEND, (event) => this._end(event));
    }
  }
  _eventIsPointerPenTouch(event) {
    return this._supportPointerEvents && (event.pointerType === POINTER_TYPE_PEN || event.pointerType === POINTER_TYPE_TOUCH);
  }
  static isSupported() {
    return "ontouchstart" in document.documentElement || navigator.maxTouchPoints > 0;
  }
};

// node_modules/bootstrap/js/dist/drawer.js
var NAME3 = "drawer";
var EVENT_KEY3 = `.bs.drawer`;
var DATA_API_KEY2 = ".data-api";
var EVENT_LOAD_DATA_API = `load${EVENT_KEY3}${DATA_API_KEY2}`;
var EVENT_HIDDEN2 = `hidden${EVENT_KEY3}`;
var EVENT_RESIZE = `resize${EVENT_KEY3}`;
var EVENT_CLICK_DATA_API2 = `click${EVENT_KEY3}${DATA_API_KEY2}`;
var SELECTOR_DATA_TOGGLE2 = '[data-bs-toggle="drawer"]';
var SELECTOR_OPEN = 'dialog.drawer[open], dialog[open][class*="\\:drawer"]';
var Default3 = {
  backdrop: true,
  keyboard: true,
  scroll: false
};
var DefaultType3 = {
  backdrop: "(boolean|string)",
  keyboard: "boolean",
  scroll: "boolean"
};
var Drawer = class extends DialogBase {
  constructor(element, config) {
    super(element, config);
    this._swipeHelper = null;
    this._resizeObserver = null;
    this._resizeFrame = null;
    this._initResizeObserver();
  }
  static get Default() {
    return Default3;
  }
  static get DefaultType() {
    return DefaultType3;
  }
  static get NAME() {
    return NAME3;
  }
  dispose() {
    if (this._swipeHelper) this._swipeHelper.dispose();
    if (this._resizeObserver) this._resizeObserver.disconnect();
    if (this._resizeFrame !== null) cancelAnimationFrame(this._resizeFrame);
    super.dispose();
  }
  _getShowOptions() {
    return {
      modal: Boolean(this._config.backdrop) || !this._config.scroll,
      preventBodyScroll: !this._config.scroll
    };
  }
  _onBeforeShow() {
    this._initSwipe();
  }
  _getInstantClassName() {
    return "drawer-instant";
  }
  _getStaticClassName() {
    return "drawer-static";
  }
  _initResizeObserver() {
    const navbar = this._element.closest(".navbar");
    if (!navbar || typeof ResizeObserver === "undefined") return;
    this._resizeObserver = new ResizeObserver(() => {
      if (this._resizeFrame !== null) cancelAnimationFrame(this._resizeFrame);
      this._resizeFrame = requestAnimationFrame(() => {
        this._resizeFrame = null;
        if (this._element.open && getComputedStyle(this._element).position !== "fixed") this.hide();
      });
    });
    this._resizeObserver.observe(navbar);
  }
  _initSwipe() {
    if (this._swipeHelper || !Swipe.isSupported()) return;
    const swipeConfig = {};
    const element = this._element;
    if (element.classList.contains("drawer-bottom")) swipeConfig.downCallback = () => this.hide();
    else if (element.classList.contains("drawer-top")) swipeConfig.upCallback = () => this.hide();
    else if (element.classList.contains("drawer-end")) if (isRTL()) swipeConfig.leftCallback = () => this.hide();
    else swipeConfig.rightCallback = () => this.hide();
    else if (isRTL()) swipeConfig.rightCallback = () => this.hide();
    else swipeConfig.leftCallback = () => this.hide();
    this._swipeHelper = new Swipe(element, swipeConfig);
  }
};
EventHandler.on(document, EVENT_CLICK_DATA_API2, SELECTOR_DATA_TOGGLE2, function(event) {
  const target = SelectorEngine.getElementFromSelector(this);
  if (["A", "AREA"].includes(this.tagName)) event.preventDefault();
  if (isDisabled(this)) return;
  EventHandler.one(target, EVENT_HIDDEN2, () => {
    if (isVisible(this)) this.focus({ preventScroll: true });
  });
  const alreadyOpen = SelectorEngine.findOne("dialog.drawer[open]");
  if (alreadyOpen && alreadyOpen !== target) Drawer.getInstance(alreadyOpen).hide();
  Drawer.getOrCreateInstance(target).toggle(this);
});
EventHandler.on(window, EVENT_LOAD_DATA_API, () => {
  for (const selector of SelectorEngine.find("dialog.drawer[open]")) Drawer.getOrCreateInstance(selector).show();
});
EventHandler.on(window, EVENT_RESIZE, () => {
  for (const element of SelectorEngine.find(SELECTOR_OPEN)) if (getComputedStyle(element).position !== "fixed") Drawer.getOrCreateInstance(element).hide();
});
enableDismissTrigger(Drawer);

// node_modules/@floating-ui/utils/dist/floating-ui.utils.mjs
var min = Math.min;
var max = Math.max;
var round = Math.round;
var floor = Math.floor;
var createCoords = (v) => ({
  x: v,
  y: v
});
var oppositeSideMap = {
  left: "right",
  right: "left",
  bottom: "top",
  top: "bottom"
};
function clamp(start, value, end) {
  return max(start, min(value, end));
}
function evaluate(value, param) {
  return typeof value === "function" ? value(param) : value;
}
function getSide(placement) {
  return placement.split("-")[0];
}
function getAlignment(placement) {
  return placement.split("-")[1];
}
function getOppositeAxis(axis) {
  return axis === "x" ? "y" : "x";
}
function getAxisLength(axis) {
  return axis === "y" ? "height" : "width";
}
function getSideAxis(placement) {
  const firstChar = placement[0];
  return firstChar === "t" || firstChar === "b" ? "y" : "x";
}
function getAlignmentAxis(placement) {
  return getOppositeAxis(getSideAxis(placement));
}
function getAlignmentSides(placement, rects, rtl) {
  if (rtl === void 0) {
    rtl = false;
  }
  const alignment = getAlignment(placement);
  const alignmentAxis = getAlignmentAxis(placement);
  const length = getAxisLength(alignmentAxis);
  let mainAlignmentSide = alignmentAxis === "x" ? alignment === (rtl ? "end" : "start") ? "right" : "left" : alignment === "start" ? "bottom" : "top";
  if (rects.reference[length] > rects.floating[length]) {
    mainAlignmentSide = getOppositePlacement(mainAlignmentSide);
  }
  return [mainAlignmentSide, getOppositePlacement(mainAlignmentSide)];
}
function getExpandedPlacements(placement) {
  const oppositePlacement = getOppositePlacement(placement);
  return [getOppositeAlignmentPlacement(placement), oppositePlacement, getOppositeAlignmentPlacement(oppositePlacement)];
}
function getOppositeAlignmentPlacement(placement) {
  return placement.includes("start") ? placement.replace("start", "end") : placement.replace("end", "start");
}
var lrPlacement = ["left", "right"];
var rlPlacement = ["right", "left"];
var tbPlacement = ["top", "bottom"];
var btPlacement = ["bottom", "top"];
function getSideList(side, isStart, rtl) {
  switch (side) {
    case "top":
    case "bottom":
      if (rtl) return isStart ? rlPlacement : lrPlacement;
      return isStart ? lrPlacement : rlPlacement;
    case "left":
    case "right":
      return isStart ? tbPlacement : btPlacement;
    default:
      return [];
  }
}
function getOppositeAxisPlacements(placement, flipAlignment, direction, rtl) {
  const alignment = getAlignment(placement);
  let list = getSideList(getSide(placement), direction === "start", rtl);
  if (alignment) {
    list = list.map((side) => side + "-" + alignment);
    if (flipAlignment) {
      list = list.concat(list.map(getOppositeAlignmentPlacement));
    }
  }
  return list;
}
function getOppositePlacement(placement) {
  const side = getSide(placement);
  return oppositeSideMap[side] + placement.slice(side.length);
}
function expandPaddingObject(padding) {
  var _padding$top, _padding$right, _padding$bottom, _padding$left;
  return {
    top: (_padding$top = padding.top) != null ? _padding$top : 0,
    right: (_padding$right = padding.right) != null ? _padding$right : 0,
    bottom: (_padding$bottom = padding.bottom) != null ? _padding$bottom : 0,
    left: (_padding$left = padding.left) != null ? _padding$left : 0
  };
}
function getPaddingObject(padding) {
  return typeof padding !== "number" ? expandPaddingObject(padding) : {
    top: padding,
    right: padding,
    bottom: padding,
    left: padding
  };
}
function rectToClientRect(rect) {
  const {
    x,
    y,
    width,
    height
  } = rect;
  return {
    width,
    height,
    top: y,
    left: x,
    right: x + width,
    bottom: y + height,
    x,
    y
  };
}

// node_modules/@floating-ui/core/dist/floating-ui.core.mjs
function computeCoordsFromPlacement(_ref, placement, rtl) {
  let {
    reference,
    floating
  } = _ref;
  const sideAxis = getSideAxis(placement);
  const alignmentAxis = getAlignmentAxis(placement);
  const alignLength = getAxisLength(alignmentAxis);
  const side = getSide(placement);
  const isVertical = sideAxis === "y";
  const commonX = reference.x + reference.width / 2 - floating.width / 2;
  const commonY = reference.y + reference.height / 2 - floating.height / 2;
  const commonAlign = reference[alignLength] / 2 - floating[alignLength] / 2;
  let coords;
  switch (side) {
    case "top":
      coords = {
        x: commonX,
        y: reference.y - floating.height
      };
      break;
    case "bottom":
      coords = {
        x: commonX,
        y: reference.y + reference.height
      };
      break;
    case "right":
      coords = {
        x: reference.x + reference.width,
        y: commonY
      };
      break;
    case "left":
      coords = {
        x: reference.x - floating.width,
        y: commonY
      };
      break;
    default:
      coords = {
        x: reference.x,
        y: reference.y
      };
  }
  const alignment = getAlignment(placement);
  if (alignment) {
    coords[alignmentAxis] += commonAlign * (alignment === "end" ? 1 : -1) * (rtl && isVertical ? -1 : 1);
  }
  return coords;
}
async function detectOverflow(state, options) {
  var _await$platform$isEle;
  if (options === void 0) {
    options = {};
  }
  const {
    x,
    y,
    platform: platform2,
    rects,
    elements,
    strategy
  } = state;
  const {
    boundary = "clippingAncestors",
    rootBoundary = "viewport",
    elementContext = "floating",
    altBoundary = false,
    padding = 0
  } = evaluate(options, state);
  const paddingObject = getPaddingObject(padding);
  const altContext = elementContext === "floating" ? "reference" : "floating";
  const element = elements[altBoundary ? altContext : elementContext];
  const clippingClientRect = rectToClientRect(await platform2.getClippingRect({
    element: ((_await$platform$isEle = await (platform2.isElement == null ? void 0 : platform2.isElement(element))) != null ? _await$platform$isEle : true) ? element : element.contextElement || await (platform2.getDocumentElement == null ? void 0 : platform2.getDocumentElement(elements.floating)),
    boundary,
    rootBoundary,
    strategy
  }));
  const rect = elementContext === "floating" ? {
    x,
    y,
    width: rects.floating.width,
    height: rects.floating.height
  } : rects.reference;
  const offsetParent = await (platform2.getOffsetParent == null ? void 0 : platform2.getOffsetParent(elements.floating));
  const offsetScale = await (platform2.isElement == null ? void 0 : platform2.isElement(offsetParent)) && await (platform2.getScale == null ? void 0 : platform2.getScale(offsetParent)) || {
    x: 1,
    y: 1
  };
  const elementClientRect = rectToClientRect(platform2.convertOffsetParentRelativeRectToViewportRelativeRect ? await platform2.convertOffsetParentRelativeRectToViewportRelativeRect({
    elements,
    rect,
    offsetParent,
    strategy
  }) : rect);
  return {
    top: (clippingClientRect.top - elementClientRect.top + paddingObject.top) / offsetScale.y,
    bottom: (elementClientRect.bottom - clippingClientRect.bottom + paddingObject.bottom) / offsetScale.y,
    left: (clippingClientRect.left - elementClientRect.left + paddingObject.left) / offsetScale.x,
    right: (elementClientRect.right - clippingClientRect.right + paddingObject.right) / offsetScale.x
  };
}
var MAX_RESET_COUNT = 50;
var computePosition = async (reference, floating, config) => {
  const {
    placement = "bottom",
    strategy = "absolute",
    middleware = [],
    platform: platform2
  } = config;
  const platformWithDetectOverflow = platform2.detectOverflow ? platform2 : {
    ...platform2,
    detectOverflow
  };
  const rtl = await (platform2.isRTL == null ? void 0 : platform2.isRTL(floating));
  let rects = await platform2.getElementRects({
    reference,
    floating,
    strategy
  });
  let {
    x,
    y
  } = computeCoordsFromPlacement(rects, placement, rtl);
  let statefulPlacement = placement;
  let resetCount = 0;
  const middlewareData = {};
  for (let i = 0; i < middleware.length; i++) {
    const currentMiddleware = middleware[i];
    if (!currentMiddleware) {
      continue;
    }
    const {
      name,
      fn
    } = currentMiddleware;
    const {
      x: nextX,
      y: nextY,
      data,
      reset
    } = await fn({
      x,
      y,
      initialPlacement: placement,
      placement: statefulPlacement,
      strategy,
      middlewareData,
      rects,
      platform: platformWithDetectOverflow,
      elements: {
        reference,
        floating
      }
    });
    x = nextX != null ? nextX : x;
    y = nextY != null ? nextY : y;
    middlewareData[name] = {
      ...middlewareData[name],
      ...data
    };
    if (reset && resetCount < MAX_RESET_COUNT) {
      resetCount++;
      if (typeof reset === "object") {
        if (reset.placement) {
          statefulPlacement = reset.placement;
        }
        if (reset.rects) {
          rects = reset.rects === true ? await platform2.getElementRects({
            reference,
            floating,
            strategy
          }) : reset.rects;
        }
        ({
          x,
          y
        } = computeCoordsFromPlacement(rects, statefulPlacement, rtl));
      }
      i = -1;
    }
  }
  return {
    x,
    y,
    placement: statefulPlacement,
    strategy,
    middlewareData
  };
};
var arrow = (options) => ({
  name: "arrow",
  options,
  async fn(state) {
    const {
      x,
      y,
      placement,
      rects,
      platform: platform2,
      elements,
      middlewareData
    } = state;
    const {
      element,
      padding = 0
    } = evaluate(options, state) || {};
    if (element == null) {
      return {};
    }
    const paddingObject = getPaddingObject(padding);
    const coords = {
      x,
      y
    };
    const axis = getAlignmentAxis(placement);
    const length = getAxisLength(axis);
    const arrowDimensions = await platform2.getDimensions(element);
    const isYAxis = axis === "y";
    const minProp = isYAxis ? "top" : "left";
    const maxProp = isYAxis ? "bottom" : "right";
    const clientProp = isYAxis ? "clientHeight" : "clientWidth";
    const endDiff = rects.reference[length] + rects.reference[axis] - coords[axis] - rects.floating[length];
    const startDiff = coords[axis] - rects.reference[axis];
    const arrowOffsetParent = await (platform2.getOffsetParent == null ? void 0 : platform2.getOffsetParent(element));
    let clientSize = arrowOffsetParent ? arrowOffsetParent[clientProp] : 0;
    if (!clientSize || !await (platform2.isElement == null ? void 0 : platform2.isElement(arrowOffsetParent))) {
      clientSize = elements.floating[clientProp] || rects.floating[length];
    }
    const centerToReference = endDiff / 2 - startDiff / 2;
    const largestPossiblePadding = clientSize / 2 - arrowDimensions[length] / 2 - 1;
    const minPadding = min(paddingObject[minProp], largestPossiblePadding);
    const maxPadding = min(paddingObject[maxProp], largestPossiblePadding);
    const max2 = clientSize - arrowDimensions[length] - maxPadding;
    const center = clientSize / 2 - arrowDimensions[length] / 2 + centerToReference;
    const offset3 = clamp(minPadding, center, max2);
    const shouldAddOffset = !middlewareData.arrow && getAlignment(placement) != null && center !== offset3 && rects.reference[length] / 2 - (center < minPadding ? minPadding : maxPadding) - arrowDimensions[length] / 2 < 0;
    const alignmentOffset = shouldAddOffset ? center < minPadding ? center - minPadding : center - max2 : 0;
    return {
      [axis]: coords[axis] + alignmentOffset,
      data: {
        [axis]: offset3,
        centerOffset: center - offset3 - alignmentOffset,
        ...shouldAddOffset && {
          alignmentOffset
        }
      },
      reset: shouldAddOffset
    };
  }
});
var flip = function(options) {
  if (options === void 0) {
    options = {};
  }
  return {
    name: "flip",
    options,
    async fn(state) {
      var _middlewareData$arrow, _middlewareData$flip;
      const {
        placement,
        middlewareData,
        rects,
        initialPlacement,
        platform: platform2,
        elements
      } = state;
      const {
        mainAxis: checkMainAxis = true,
        crossAxis: checkCrossAxis = true,
        fallbackPlacements: specifiedFallbackPlacements,
        fallbackStrategy = "bestFit",
        fallbackAxisSideDirection = "none",
        flipAlignment = true,
        ...detectOverflowOptions
      } = evaluate(options, state);
      if ((_middlewareData$arrow = middlewareData.arrow) != null && _middlewareData$arrow.alignmentOffset) {
        return {};
      }
      const side = getSide(placement);
      const initialSideAxis = getSideAxis(initialPlacement);
      const isBasePlacement = getSide(initialPlacement) === initialPlacement;
      const rtl = await (platform2.isRTL == null ? void 0 : platform2.isRTL(elements.floating));
      const fallbackPlacements = specifiedFallbackPlacements || (isBasePlacement || !flipAlignment ? [getOppositePlacement(initialPlacement)] : getExpandedPlacements(initialPlacement));
      const hasFallbackAxisSideDirection = fallbackAxisSideDirection !== "none";
      if (!specifiedFallbackPlacements && hasFallbackAxisSideDirection) {
        fallbackPlacements.push(...getOppositeAxisPlacements(initialPlacement, flipAlignment, fallbackAxisSideDirection, rtl));
      }
      const placements2 = [initialPlacement, ...fallbackPlacements];
      const overflow = await platform2.detectOverflow(state, detectOverflowOptions);
      const overflows = [];
      let overflowsData = ((_middlewareData$flip = middlewareData.flip) == null ? void 0 : _middlewareData$flip.overflows) || [];
      if (checkMainAxis) {
        overflows.push(overflow[side]);
      }
      if (checkCrossAxis) {
        const sides2 = getAlignmentSides(placement, rects, rtl);
        overflows.push(overflow[sides2[0]], overflow[sides2[1]]);
      }
      overflowsData = [...overflowsData, {
        placement,
        overflows
      }];
      if (!overflows.every((side2) => side2 <= 0)) {
        var _middlewareData$flip2, _overflowsData$filter;
        const nextIndex = (((_middlewareData$flip2 = middlewareData.flip) == null ? void 0 : _middlewareData$flip2.index) || 0) + 1;
        const nextPlacement = placements2[nextIndex];
        if (nextPlacement) {
          const ignoreCrossAxisOverflow = checkCrossAxis === "alignment" ? initialSideAxis !== getSideAxis(nextPlacement) : false;
          if (!ignoreCrossAxisOverflow || // We leave the current main axis only if every placement on that axis
          // overflows the main axis.
          overflowsData.every((d) => getSideAxis(d.placement) === initialSideAxis ? d.overflows[0] > 0 : true)) {
            return {
              data: {
                index: nextIndex,
                overflows: overflowsData
              },
              reset: {
                placement: nextPlacement
              }
            };
          }
        }
        let resetPlacement = (_overflowsData$filter = overflowsData.filter((d) => d.overflows[0] <= 0).sort((a, b) => a.overflows[1] - b.overflows[1])[0]) == null ? void 0 : _overflowsData$filter.placement;
        if (!resetPlacement) {
          switch (fallbackStrategy) {
            case "bestFit": {
              var _overflowsData$filter2;
              const placement2 = (_overflowsData$filter2 = overflowsData.filter((d) => {
                if (hasFallbackAxisSideDirection) {
                  const currentSideAxis = getSideAxis(d.placement);
                  return currentSideAxis === initialSideAxis || // Create a bias to the `y` side axis due to horizontal
                  // reading directions favoring greater width.
                  currentSideAxis === "y";
                }
                return true;
              }).map((d) => [d.placement, d.overflows.filter((overflow2) => overflow2 > 0).reduce((acc, overflow2) => acc + overflow2, 0)]).sort((a, b) => a[1] - b[1])[0]) == null ? void 0 : _overflowsData$filter2[0];
              if (placement2) {
                resetPlacement = placement2;
              }
              break;
            }
            case "initialPlacement":
              resetPlacement = initialPlacement;
              break;
          }
        }
        if (placement !== resetPlacement) {
          return {
            reset: {
              placement: resetPlacement
            }
          };
        }
      }
      return {};
    }
  };
};
var originSides = /* @__PURE__ */ new Set(["left", "top"]);
async function convertValueToCoords(state, options) {
  const {
    placement,
    platform: platform2,
    elements
  } = state;
  const rtl = await (platform2.isRTL == null ? void 0 : platform2.isRTL(elements.floating));
  const side = getSide(placement);
  const alignment = getAlignment(placement);
  const isVertical = getSideAxis(placement) === "y";
  const mainAxisMulti = originSides.has(side) ? -1 : 1;
  const crossAxisMulti = rtl && isVertical ? -1 : 1;
  const rawValue = evaluate(options, state);
  let {
    mainAxis,
    crossAxis,
    alignmentAxis
  } = typeof rawValue === "number" ? {
    mainAxis: rawValue,
    crossAxis: 0,
    alignmentAxis: null
  } : {
    mainAxis: rawValue.mainAxis || 0,
    crossAxis: rawValue.crossAxis || 0,
    alignmentAxis: rawValue.alignmentAxis
  };
  if (alignment && typeof alignmentAxis === "number") {
    crossAxis = alignment === "end" ? alignmentAxis * -1 : alignmentAxis;
  }
  return isVertical ? {
    x: crossAxis * crossAxisMulti,
    y: mainAxis * mainAxisMulti
  } : {
    x: mainAxis * mainAxisMulti,
    y: crossAxis * crossAxisMulti
  };
}
var offset = function(options) {
  if (options === void 0) {
    options = 0;
  }
  return {
    name: "offset",
    options,
    async fn(state) {
      var _middlewareData$offse, _middlewareData$arrow;
      const {
        x,
        y,
        placement,
        middlewareData
      } = state;
      const diffCoords = await convertValueToCoords(state, options);
      if (placement === ((_middlewareData$offse = middlewareData.offset) == null ? void 0 : _middlewareData$offse.placement) && (_middlewareData$arrow = middlewareData.arrow) != null && _middlewareData$arrow.alignmentOffset) {
        return {};
      }
      return {
        x: x + diffCoords.x,
        y: y + diffCoords.y,
        data: {
          ...diffCoords,
          placement
        }
      };
    }
  };
};
var shift = function(options) {
  if (options === void 0) {
    options = {};
  }
  return {
    name: "shift",
    options,
    async fn(state) {
      const {
        x,
        y,
        placement,
        platform: platform2
      } = state;
      const {
        mainAxis: checkMainAxis = true,
        crossAxis: checkCrossAxis = false,
        limiter = {
          fn: (_ref) => {
            let {
              x: x2,
              y: y2
            } = _ref;
            return {
              x: x2,
              y: y2
            };
          }
        },
        ...detectOverflowOptions
      } = evaluate(options, state);
      const coords = {
        x,
        y
      };
      const overflow = await platform2.detectOverflow(state, detectOverflowOptions);
      const crossAxis = getSideAxis(placement);
      const mainAxis = getOppositeAxis(crossAxis);
      let mainAxisCoord = coords[mainAxis];
      let crossAxisCoord = coords[crossAxis];
      const clampCoord = (axis, coord) => clamp(coord + overflow[axis === "y" ? "top" : "left"], coord, coord - overflow[axis === "y" ? "bottom" : "right"]);
      if (checkMainAxis) {
        mainAxisCoord = clampCoord(mainAxis, mainAxisCoord);
      }
      if (checkCrossAxis) {
        crossAxisCoord = clampCoord(crossAxis, crossAxisCoord);
      }
      const limitedCoords = limiter.fn({
        ...state,
        [mainAxis]: mainAxisCoord,
        [crossAxis]: crossAxisCoord
      });
      return {
        ...limitedCoords,
        data: {
          x: limitedCoords.x - x,
          y: limitedCoords.y - y,
          enabled: {
            [mainAxis]: checkMainAxis,
            [crossAxis]: checkCrossAxis
          }
        }
      };
    }
  };
};

// node_modules/@floating-ui/utils/dist/floating-ui.utils.dom.mjs
function hasWindow() {
  return typeof window !== "undefined";
}
function getNodeName(node) {
  if (isNode(node)) {
    return (node.nodeName || "").toLowerCase();
  }
  return "#document";
}
function getWindow(node) {
  var _node$ownerDocument;
  return (node == null || (_node$ownerDocument = node.ownerDocument) == null ? void 0 : _node$ownerDocument.defaultView) || window;
}
function getDocumentElement(node) {
  var _ref;
  return (_ref = (isNode(node) ? node.ownerDocument : node.document) || window.document) == null ? void 0 : _ref.documentElement;
}
function isNode(value) {
  if (!hasWindow()) {
    return false;
  }
  return value instanceof Node || value instanceof getWindow(value).Node;
}
function isElement2(value) {
  if (!hasWindow()) {
    return false;
  }
  return value instanceof Element || value instanceof getWindow(value).Element;
}
function isHTMLElement(value) {
  if (!hasWindow()) {
    return false;
  }
  return value instanceof HTMLElement || value instanceof getWindow(value).HTMLElement;
}
function isShadowRoot(value) {
  if (!hasWindow() || typeof ShadowRoot === "undefined") {
    return false;
  }
  return value instanceof ShadowRoot || value instanceof getWindow(value).ShadowRoot;
}
function isOverflowElement(element) {
  const {
    overflow,
    overflowX,
    overflowY,
    display
  } = getComputedStyle2(element);
  return /auto|scroll|overlay|hidden|clip/.test(overflow + overflowY + overflowX) && display !== "inline" && display !== "contents";
}
function isTableElement(element) {
  return /^(table|td|th)$/.test(getNodeName(element));
}
function isTopLayer(element) {
  try {
    if (element.matches(":popover-open")) {
      return true;
    }
  } catch (_e) {
  }
  try {
    return element.matches(":modal");
  } catch (_e) {
    return false;
  }
}
var willChangeRe = /transform|translate|scale|rotate|perspective|filter/;
var containRe = /paint|layout|strict|content/;
var isNotNone = (value) => !!value && value !== "none";
var isWebKitValue;
function isContainingBlock(elementOrCss) {
  const css = isElement2(elementOrCss) ? getComputedStyle2(elementOrCss) : elementOrCss;
  return isNotNone(css.transform) || isNotNone(css.translate) || isNotNone(css.scale) || isNotNone(css.rotate) || isNotNone(css.perspective) || !isWebKit() && (isNotNone(css.backdropFilter) || isNotNone(css.filter)) || willChangeRe.test(css.willChange || "") || containRe.test(css.contain || "");
}
function getContainingBlock(element) {
  let currentNode = getParentNode(element);
  while (isHTMLElement(currentNode) && !isLastTraversableNode(currentNode)) {
    if (isContainingBlock(currentNode)) {
      return currentNode;
    } else if (isTopLayer(currentNode)) {
      return null;
    }
    currentNode = getParentNode(currentNode);
  }
  return null;
}
function isWebKit() {
  if (isWebKitValue == null) {
    isWebKitValue = typeof CSS !== "undefined" && CSS.supports && CSS.supports("-webkit-backdrop-filter", "none");
  }
  return isWebKitValue;
}
function isLastTraversableNode(node) {
  return /^(html|body|#document)$/.test(getNodeName(node));
}
function getComputedStyle2(element) {
  return getWindow(element).getComputedStyle(element);
}
function getNodeScroll(element) {
  if (isElement2(element)) {
    return {
      scrollLeft: element.scrollLeft,
      scrollTop: element.scrollTop
    };
  }
  return {
    scrollLeft: element.scrollX,
    scrollTop: element.scrollY
  };
}
function getParentNode(node) {
  if (getNodeName(node) === "html") {
    return node;
  }
  const result = (
    // Step into the shadow DOM of the parent of a slotted node.
    node.assignedSlot || // DOM Element detected.
    node.parentNode || // ShadowRoot detected.
    isShadowRoot(node) && node.host || // Fallback.
    getDocumentElement(node)
  );
  return isShadowRoot(result) ? result.host : result;
}
function getNearestOverflowAncestor(node) {
  const parentNode = getParentNode(node);
  if (isLastTraversableNode(parentNode)) {
    return (node.ownerDocument || node).body;
  }
  if (isHTMLElement(parentNode) && isOverflowElement(parentNode)) {
    return parentNode;
  }
  return getNearestOverflowAncestor(parentNode);
}
function getOverflowAncestors(node, list, traverseIframes) {
  var _node$ownerDocument2;
  if (list === void 0) {
    list = [];
  }
  if (traverseIframes === void 0) {
    traverseIframes = true;
  }
  const scrollableAncestor = getNearestOverflowAncestor(node);
  const isBody = scrollableAncestor === ((_node$ownerDocument2 = node.ownerDocument) == null ? void 0 : _node$ownerDocument2.body);
  const win = getWindow(scrollableAncestor);
  if (isBody) {
    const frameElement = getFrameElement(win);
    return list.concat(win, win.visualViewport || [], isOverflowElement(scrollableAncestor) ? scrollableAncestor : [], frameElement && traverseIframes ? getOverflowAncestors(frameElement) : []);
  } else {
    return list.concat(scrollableAncestor, getOverflowAncestors(scrollableAncestor, [], traverseIframes));
  }
}
function getFrameElement(win) {
  return win.parent && Object.getPrototypeOf(win.parent) ? win.frameElement : null;
}

// node_modules/@floating-ui/dom/dist/floating-ui.dom.mjs
function getCssDimensions(element) {
  const css = getComputedStyle2(element);
  let width = parseFloat(css.width) || 0;
  let height = parseFloat(css.height) || 0;
  const hasOffset = isHTMLElement(element);
  const offsetWidth = hasOffset ? element.offsetWidth : width;
  const offsetHeight = hasOffset ? element.offsetHeight : height;
  const shouldFallback = round(width) !== offsetWidth || round(height) !== offsetHeight;
  if (shouldFallback) {
    width = offsetWidth;
    height = offsetHeight;
  }
  return {
    width,
    height,
    $: shouldFallback
  };
}
function unwrapElement(element) {
  return !isElement2(element) ? element.contextElement : element;
}
function getScale(element) {
  const domElement = unwrapElement(element);
  if (!isHTMLElement(domElement)) {
    return createCoords(1);
  }
  const rect = domElement.getBoundingClientRect();
  const {
    width,
    height,
    $
  } = getCssDimensions(domElement);
  let x = ($ ? round(rect.width) : rect.width) / width;
  let y = ($ ? round(rect.height) : rect.height) / height;
  if (!x || !Number.isFinite(x)) {
    x = 1;
  }
  if (!y || !Number.isFinite(y)) {
    y = 1;
  }
  return {
    x,
    y
  };
}
var noOffsets = /* @__PURE__ */ createCoords(0);
function getVisualOffsets(element) {
  const win = getWindow(element);
  if (!isWebKit() || !win.visualViewport) {
    return noOffsets;
  }
  return {
    x: win.visualViewport.offsetLeft,
    y: win.visualViewport.offsetTop
  };
}
function shouldAddVisualOffsets(element, isFixed, floatingOffsetParent) {
  if (isFixed === void 0) {
    isFixed = false;
  }
  if (!floatingOffsetParent || isFixed && floatingOffsetParent !== getWindow(element)) {
    return false;
  }
  return isFixed;
}
function getBoundingClientRect(element, includeScale, isFixedStrategy, offsetParent) {
  if (includeScale === void 0) {
    includeScale = false;
  }
  if (isFixedStrategy === void 0) {
    isFixedStrategy = false;
  }
  const clientRect = element.getBoundingClientRect();
  const domElement = unwrapElement(element);
  let scale = createCoords(1);
  if (includeScale) {
    if (offsetParent) {
      if (isElement2(offsetParent)) {
        scale = getScale(offsetParent);
      }
    } else {
      scale = getScale(element);
    }
  }
  const visualOffsets = shouldAddVisualOffsets(domElement, isFixedStrategy, offsetParent) ? getVisualOffsets(domElement) : createCoords(0);
  let x = (clientRect.left + visualOffsets.x) / scale.x;
  let y = (clientRect.top + visualOffsets.y) / scale.y;
  let width = clientRect.width / scale.x;
  let height = clientRect.height / scale.y;
  if (domElement) {
    const win = getWindow(domElement);
    const offsetWin = offsetParent && isElement2(offsetParent) ? getWindow(offsetParent) : offsetParent;
    let currentWin = win;
    let currentIFrame = getFrameElement(currentWin);
    while (currentIFrame && offsetParent && offsetWin !== currentWin) {
      const iframeScale = getScale(currentIFrame);
      const iframeRect = currentIFrame.getBoundingClientRect();
      const css = getComputedStyle2(currentIFrame);
      const left = iframeRect.left + (currentIFrame.clientLeft + parseFloat(css.paddingLeft)) * iframeScale.x;
      const top = iframeRect.top + (currentIFrame.clientTop + parseFloat(css.paddingTop)) * iframeScale.y;
      x *= iframeScale.x;
      y *= iframeScale.y;
      width *= iframeScale.x;
      height *= iframeScale.y;
      x += left;
      y += top;
      currentWin = getWindow(currentIFrame);
      currentIFrame = getFrameElement(currentWin);
    }
  }
  return rectToClientRect({
    width,
    height,
    x,
    y
  });
}
function getWindowScrollBarX(element, rect) {
  const leftScroll = getNodeScroll(element).scrollLeft;
  if (!rect) {
    return getBoundingClientRect(getDocumentElement(element)).left + leftScroll;
  }
  return rect.left + leftScroll;
}
function getHTMLOffset(documentElement, scroll) {
  const htmlRect = documentElement.getBoundingClientRect();
  const x = htmlRect.left + scroll.scrollLeft - getWindowScrollBarX(documentElement, htmlRect);
  const y = htmlRect.top + scroll.scrollTop;
  return {
    x,
    y
  };
}
function convertOffsetParentRelativeRectToViewportRelativeRect(_ref) {
  let {
    elements,
    rect,
    offsetParent,
    strategy
  } = _ref;
  const isFixed = strategy === "fixed";
  const documentElement = getDocumentElement(offsetParent);
  const topLayer = elements ? isTopLayer(elements.floating) : false;
  if (offsetParent === documentElement || topLayer && isFixed) {
    return rect;
  }
  let scroll = {
    scrollLeft: 0,
    scrollTop: 0
  };
  let scale = createCoords(1);
  const offsets = createCoords(0);
  const isOffsetParentAnElement = isHTMLElement(offsetParent);
  if (isOffsetParentAnElement || !isOffsetParentAnElement && !isFixed) {
    if (getNodeName(offsetParent) !== "body" || isOverflowElement(documentElement)) {
      scroll = getNodeScroll(offsetParent);
    }
    if (isOffsetParentAnElement) {
      const offsetRect = getBoundingClientRect(offsetParent);
      scale = getScale(offsetParent);
      offsets.x = offsetRect.x + offsetParent.clientLeft;
      offsets.y = offsetRect.y + offsetParent.clientTop;
    }
  }
  const htmlOffset = documentElement && !isOffsetParentAnElement && !isFixed ? getHTMLOffset(documentElement, scroll) : createCoords(0);
  return {
    width: rect.width * scale.x,
    height: rect.height * scale.y,
    x: rect.x * scale.x - scroll.scrollLeft * scale.x + offsets.x + htmlOffset.x,
    y: rect.y * scale.y - scroll.scrollTop * scale.y + offsets.y + htmlOffset.y
  };
}
function getClientRects(element) {
  return Array.from(element.getClientRects());
}
function getDocumentRect(element) {
  const html = getDocumentElement(element);
  const scroll = getNodeScroll(element);
  const body = element.ownerDocument.body;
  const width = max(html.scrollWidth, html.clientWidth, body.scrollWidth, body.clientWidth);
  const height = max(html.scrollHeight, html.clientHeight, body.scrollHeight, body.clientHeight);
  let x = -scroll.scrollLeft + getWindowScrollBarX(element);
  const y = -scroll.scrollTop;
  if (getComputedStyle2(body).direction === "rtl") {
    x += max(html.clientWidth, body.clientWidth) - width;
  }
  return {
    width,
    height,
    x,
    y
  };
}
var SCROLLBAR_MAX = 25;
function getViewportRect(element, strategy) {
  const win = getWindow(element);
  const html = getDocumentElement(element);
  const visualViewport = win.visualViewport;
  let width = html.clientWidth;
  let height = html.clientHeight;
  let x = 0;
  let y = 0;
  if (visualViewport) {
    width = visualViewport.width;
    height = visualViewport.height;
    const visualViewportBased = isWebKit();
    if (!visualViewportBased || visualViewportBased && strategy === "fixed") {
      x = visualViewport.offsetLeft;
      y = visualViewport.offsetTop;
    }
  }
  const windowScrollbarX = getWindowScrollBarX(html);
  if (windowScrollbarX <= 0) {
    const doc = html.ownerDocument;
    const body = doc.body;
    const bodyStyles = getComputedStyle(body);
    const bodyMarginInline = doc.compatMode === "CSS1Compat" ? parseFloat(bodyStyles.marginLeft) + parseFloat(bodyStyles.marginRight) || 0 : 0;
    const clippingStableScrollbarWidth = Math.abs(html.clientWidth - body.clientWidth - bodyMarginInline);
    if (clippingStableScrollbarWidth <= SCROLLBAR_MAX) {
      width -= clippingStableScrollbarWidth;
    }
  } else if (windowScrollbarX <= SCROLLBAR_MAX) {
    width += windowScrollbarX;
  }
  return {
    width,
    height,
    x,
    y
  };
}
function getInnerBoundingClientRect(element, strategy) {
  const clientRect = getBoundingClientRect(element, true, strategy === "fixed");
  const top = clientRect.top + element.clientTop;
  const left = clientRect.left + element.clientLeft;
  const scale = isHTMLElement(element) ? getScale(element) : createCoords(1);
  const width = element.clientWidth * scale.x;
  const height = element.clientHeight * scale.y;
  const x = left * scale.x;
  const y = top * scale.y;
  return {
    width,
    height,
    x,
    y
  };
}
function getClientRectFromClippingAncestor(element, clippingAncestor, strategy) {
  let rect;
  if (clippingAncestor === "viewport") {
    rect = getViewportRect(element, strategy);
  } else if (clippingAncestor === "document") {
    rect = getDocumentRect(getDocumentElement(element));
  } else if (isElement2(clippingAncestor)) {
    rect = getInnerBoundingClientRect(clippingAncestor, strategy);
  } else {
    const visualOffsets = getVisualOffsets(element);
    rect = {
      x: clippingAncestor.x - visualOffsets.x,
      y: clippingAncestor.y - visualOffsets.y,
      width: clippingAncestor.width,
      height: clippingAncestor.height
    };
  }
  return rectToClientRect(rect);
}
function hasFixedPositionAncestor(element, stopNode) {
  const parentNode = getParentNode(element);
  if (parentNode === stopNode || !isElement2(parentNode) || isLastTraversableNode(parentNode)) {
    return false;
  }
  return getComputedStyle2(parentNode).position === "fixed" || hasFixedPositionAncestor(parentNode, stopNode);
}
function getClippingElementAncestors(element, cache) {
  const cachedResult = cache.get(element);
  if (cachedResult) {
    return cachedResult;
  }
  let result = getOverflowAncestors(element, [], false).filter((el) => isElement2(el) && getNodeName(el) !== "body");
  let currentContainingBlockComputedStyle = null;
  const elementIsFixed = getComputedStyle2(element).position === "fixed";
  let currentNode = elementIsFixed ? getParentNode(element) : element;
  while (isElement2(currentNode) && !isLastTraversableNode(currentNode)) {
    const computedStyle = getComputedStyle2(currentNode);
    const currentNodeIsContaining = isContainingBlock(currentNode);
    if (!currentNodeIsContaining && computedStyle.position === "fixed") {
      currentContainingBlockComputedStyle = null;
    }
    const shouldDropCurrentNode = elementIsFixed ? !currentNodeIsContaining && !currentContainingBlockComputedStyle : !currentNodeIsContaining && computedStyle.position === "static" && !!currentContainingBlockComputedStyle && (currentContainingBlockComputedStyle.position === "absolute" || currentContainingBlockComputedStyle.position === "fixed") || isOverflowElement(currentNode) && !currentNodeIsContaining && hasFixedPositionAncestor(element, currentNode);
    if (shouldDropCurrentNode) {
      result = result.filter((ancestor) => ancestor !== currentNode);
    } else {
      currentContainingBlockComputedStyle = computedStyle;
    }
    currentNode = getParentNode(currentNode);
  }
  cache.set(element, result);
  return result;
}
function getClippingRect(_ref) {
  let {
    element,
    boundary,
    rootBoundary,
    strategy
  } = _ref;
  const elementClippingAncestors = boundary === "clippingAncestors" ? isTopLayer(element) ? [] : getClippingElementAncestors(element, this._c) : [].concat(boundary);
  const clippingAncestors = [...elementClippingAncestors, rootBoundary];
  const firstRect = getClientRectFromClippingAncestor(element, clippingAncestors[0], strategy);
  let top = firstRect.top;
  let right = firstRect.right;
  let bottom = firstRect.bottom;
  let left = firstRect.left;
  for (let i = 1; i < clippingAncestors.length; i++) {
    const rect = getClientRectFromClippingAncestor(element, clippingAncestors[i], strategy);
    top = max(rect.top, top);
    right = min(rect.right, right);
    bottom = min(rect.bottom, bottom);
    left = max(rect.left, left);
  }
  return {
    width: right - left,
    height: bottom - top,
    x: left,
    y: top
  };
}
function getDimensions(element) {
  const {
    width,
    height
  } = getCssDimensions(element);
  return {
    width,
    height
  };
}
function getRectRelativeToOffsetParent(element, offsetParent, strategy) {
  const isOffsetParentAnElement = isHTMLElement(offsetParent);
  const documentElement = getDocumentElement(offsetParent);
  const isFixed = strategy === "fixed";
  const rect = getBoundingClientRect(element, true, isFixed, offsetParent);
  let scroll = {
    scrollLeft: 0,
    scrollTop: 0
  };
  const offsets = createCoords(0);
  function setLeftRTLScrollbarOffset() {
    offsets.x = getWindowScrollBarX(documentElement);
  }
  if (isOffsetParentAnElement || !isOffsetParentAnElement && !isFixed) {
    if (getNodeName(offsetParent) !== "body" || isOverflowElement(documentElement)) {
      scroll = getNodeScroll(offsetParent);
    }
    if (isOffsetParentAnElement) {
      const offsetRect = getBoundingClientRect(offsetParent, true, isFixed, offsetParent);
      offsets.x = offsetRect.x + offsetParent.clientLeft;
      offsets.y = offsetRect.y + offsetParent.clientTop;
    } else if (documentElement) {
      setLeftRTLScrollbarOffset();
    }
  }
  if (isFixed && !isOffsetParentAnElement && documentElement) {
    setLeftRTLScrollbarOffset();
  }
  const htmlOffset = documentElement && !isOffsetParentAnElement && !isFixed ? getHTMLOffset(documentElement, scroll) : createCoords(0);
  const x = rect.left + scroll.scrollLeft - offsets.x - htmlOffset.x;
  const y = rect.top + scroll.scrollTop - offsets.y - htmlOffset.y;
  return {
    x,
    y,
    width: rect.width,
    height: rect.height
  };
}
function isStaticPositioned(element) {
  return getComputedStyle2(element).position === "static";
}
function getTrueOffsetParent(element, polyfill) {
  if (!isHTMLElement(element) || getComputedStyle2(element).position === "fixed") {
    return null;
  }
  if (polyfill) {
    return polyfill(element);
  }
  let rawOffsetParent = element.offsetParent;
  if (getDocumentElement(element) === rawOffsetParent) {
    rawOffsetParent = rawOffsetParent.ownerDocument.body;
  }
  return rawOffsetParent;
}
function getOffsetParent(element, polyfill) {
  const win = getWindow(element);
  if (isTopLayer(element)) {
    return win;
  }
  if (!isHTMLElement(element)) {
    let svgOffsetParent = getParentNode(element);
    while (svgOffsetParent && !isLastTraversableNode(svgOffsetParent)) {
      if (isElement2(svgOffsetParent) && !isStaticPositioned(svgOffsetParent)) {
        return svgOffsetParent;
      }
      svgOffsetParent = getParentNode(svgOffsetParent);
    }
    return win;
  }
  let offsetParent = getTrueOffsetParent(element, polyfill);
  while (offsetParent && isTableElement(offsetParent) && isStaticPositioned(offsetParent)) {
    offsetParent = getTrueOffsetParent(offsetParent, polyfill);
  }
  if (offsetParent && isLastTraversableNode(offsetParent) && isStaticPositioned(offsetParent) && !isContainingBlock(offsetParent)) {
    return win;
  }
  return offsetParent || getContainingBlock(element) || win;
}
var getElementRects = async function(data) {
  const getOffsetParentFn = this.getOffsetParent || getOffsetParent;
  const getDimensionsFn = this.getDimensions;
  const floatingDimensions = await getDimensionsFn(data.floating);
  return {
    reference: getRectRelativeToOffsetParent(data.reference, await getOffsetParentFn(data.floating), data.strategy),
    floating: {
      x: 0,
      y: 0,
      width: floatingDimensions.width,
      height: floatingDimensions.height
    }
  };
};
function isRTL2(element) {
  return getComputedStyle2(element).direction === "rtl";
}
var platform = {
  convertOffsetParentRelativeRectToViewportRelativeRect,
  getDocumentElement,
  getClippingRect,
  getOffsetParent,
  getElementRects,
  getClientRects,
  getDimensions,
  getScale,
  isElement: isElement2,
  isRTL: isRTL2
};
function rectsAreEqual(a, b) {
  return a.x === b.x && a.y === b.y && a.width === b.width && a.height === b.height;
}
function observeMove(element, onMove) {
  let io = null;
  let timeoutId;
  const root = getDocumentElement(element);
  function cleanup() {
    var _io;
    clearTimeout(timeoutId);
    (_io = io) == null || _io.disconnect();
    io = null;
  }
  function refresh(skip, threshold) {
    if (skip === void 0) {
      skip = false;
    }
    if (threshold === void 0) {
      threshold = 1;
    }
    cleanup();
    const elementRectForRootMargin = element.getBoundingClientRect();
    const {
      left,
      top,
      width,
      height
    } = elementRectForRootMargin;
    if (!skip) {
      onMove();
    }
    if (!width || !height) {
      return;
    }
    const insetTop = floor(top);
    const insetRight = floor(root.clientWidth - (left + width));
    const insetBottom = floor(root.clientHeight - (top + height));
    const insetLeft = floor(left);
    const rootMargin = -insetTop + "px " + -insetRight + "px " + -insetBottom + "px " + -insetLeft + "px";
    const options = {
      rootMargin,
      threshold: max(0, min(1, threshold)) || 1
    };
    let isFirstUpdate = true;
    function handleObserve(entries) {
      const ratio = entries[0].intersectionRatio;
      if (ratio !== threshold) {
        if (!isFirstUpdate) {
          return refresh();
        }
        if (!ratio) {
          timeoutId = setTimeout(() => {
            refresh(false, 1e-7);
          }, 1e3);
        } else {
          refresh(false, ratio);
        }
      }
      if (ratio === 1 && !rectsAreEqual(elementRectForRootMargin, element.getBoundingClientRect())) {
        refresh();
      }
      isFirstUpdate = false;
    }
    try {
      io = new IntersectionObserver(handleObserve, {
        ...options,
        // Handle <iframe>s
        root: root.ownerDocument
      });
    } catch (_e) {
      io = new IntersectionObserver(handleObserve, options);
    }
    io.observe(element);
  }
  refresh(true);
  return cleanup;
}
function autoUpdate(reference, floating, update, options) {
  if (options === void 0) {
    options = {};
  }
  const {
    ancestorScroll = true,
    ancestorResize = true,
    elementResize = typeof ResizeObserver === "function",
    layoutShift = typeof IntersectionObserver === "function",
    animationFrame = false
  } = options;
  const referenceEl = unwrapElement(reference);
  const ancestors = ancestorScroll || ancestorResize ? [...referenceEl ? getOverflowAncestors(referenceEl) : [], ...floating ? getOverflowAncestors(floating) : []] : [];
  ancestors.forEach((ancestor) => {
    ancestorScroll && ancestor.addEventListener("scroll", update, {
      passive: true
    });
    ancestorResize && ancestor.addEventListener("resize", update);
  });
  const cleanupIo = referenceEl && layoutShift ? observeMove(referenceEl, update) : null;
  let reobserveFrame = -1;
  let resizeObserver = null;
  if (elementResize) {
    resizeObserver = new ResizeObserver((_ref) => {
      let [firstEntry] = _ref;
      if (firstEntry && firstEntry.target === referenceEl && resizeObserver && floating) {
        resizeObserver.unobserve(floating);
        cancelAnimationFrame(reobserveFrame);
        reobserveFrame = requestAnimationFrame(() => {
          var _resizeObserver;
          (_resizeObserver = resizeObserver) == null || _resizeObserver.observe(floating);
        });
      }
      update();
    });
    if (referenceEl && !animationFrame) {
      resizeObserver.observe(referenceEl);
    }
    if (floating) {
      resizeObserver.observe(floating);
    }
  }
  let frameId;
  let prevRefRect = animationFrame ? getBoundingClientRect(reference) : null;
  if (animationFrame) {
    frameLoop();
  }
  function frameLoop() {
    const nextRefRect = getBoundingClientRect(reference);
    if (prevRefRect && !rectsAreEqual(prevRefRect, nextRefRect)) {
      update();
    }
    prevRefRect = nextRefRect;
    frameId = requestAnimationFrame(frameLoop);
  }
  update();
  return () => {
    var _resizeObserver2;
    ancestors.forEach((ancestor) => {
      ancestorScroll && ancestor.removeEventListener("scroll", update);
      ancestorResize && ancestor.removeEventListener("resize", update);
    });
    cleanupIo == null || cleanupIo();
    (_resizeObserver2 = resizeObserver) == null || _resizeObserver2.disconnect();
    resizeObserver = null;
    if (animationFrame) {
      cancelAnimationFrame(frameId);
    }
  };
}
var offset2 = offset;
var shift2 = shift;
var flip2 = flip;
var arrow2 = arrow;
var computePosition2 = (reference, floating, options) => {
  const cache = /* @__PURE__ */ new Map();
  const mergedOptions = {
    platform,
    ...options
  };
  const platformWithCache = {
    ...mergedOptions.platform,
    _c: cache
  };
  return computePosition(reference, floating, {
    ...mergedOptions,
    platform: platformWithCache
  });
};

// node_modules/bootstrap/js/dist/util/floating-ui.js
var BREAKPOINTS = {
  sm: 576,
  md: 768,
  lg: 1024,
  xl: 1280,
  "2xl": 1536
};
var parseResponsivePlacement = (placementString, defaultPlacement = "bottom") => {
  if (!placementString || !placementString.includes(":")) return null;
  const parts = placementString.split(/\s+/);
  const placements2 = { xs: defaultPlacement };
  for (const part of parts) if (part.includes(":")) {
    const [breakpoint, placement] = part.split(":");
    if (BREAKPOINTS[breakpoint] !== void 0) placements2[breakpoint] = placement;
  } else placements2.xs = part;
  return placements2;
};
var getResponsivePlacement = (responsivePlacements, defaultPlacement = "bottom") => {
  if (!responsivePlacements) return defaultPlacement;
  const viewportWidth = window.innerWidth;
  let activePlacement = responsivePlacements.xs || defaultPlacement;
  for (const breakpoint of [
    "sm",
    "md",
    "lg",
    "xl",
    "2xl"
  ]) if (viewportWidth >= BREAKPOINTS[breakpoint] && responsivePlacements[breakpoint]) activePlacement = responsivePlacements[breakpoint];
  return activePlacement;
};
var createBreakpointListeners = (callback) => {
  const listeners = [];
  for (const breakpoint of Object.keys(BREAKPOINTS)) {
    const minWidth = BREAKPOINTS[breakpoint];
    const mql = window.matchMedia(`(min-width: ${minWidth}px)`);
    mql.addEventListener("change", callback);
    listeners.push({
      mql,
      handler: callback
    });
  }
  return listeners;
};
var disposeBreakpointListeners = (listeners) => {
  for (const { mql, handler } of listeners) mql.removeEventListener("change", handler);
};
var toFloatingOffset = (value) => {
  return Array.isArray(value) ? {
    mainAxis: value[1] || 0,
    crossAxis: value[0] || 0
  } : value;
};

// node_modules/bootstrap/js/dist/menu.js
var NAME4 = "menu";
var EVENT_KEY4 = `.bs.menu`;
var DATA_API_KEY3 = ".data-api";
var ESCAPE_KEY = "Escape";
var TAB_KEY = "Tab";
var ARROW_UP_KEY = "ArrowUp";
var ARROW_DOWN_KEY = "ArrowDown";
var ARROW_LEFT_KEY = "ArrowLeft";
var ARROW_RIGHT_KEY = "ArrowRight";
var HOME_KEY = "Home";
var END_KEY = "End";
var ENTER_KEY = "Enter";
var SPACE_KEY = " ";
var RIGHT_MOUSE_BUTTON = 2;
var SUBMENU_CLOSE_DELAY = 100;
var EVENT_HIDE = `hide${EVENT_KEY4}`;
var EVENT_HIDDEN3 = `hidden${EVENT_KEY4}`;
var EVENT_SHOW2 = `show${EVENT_KEY4}`;
var EVENT_SHOWN = `shown${EVENT_KEY4}`;
var EVENT_CLICK_DATA_API3 = `click${EVENT_KEY4}${DATA_API_KEY3}`;
var EVENT_KEYDOWN_DATA_API = `keydown${EVENT_KEY4}${DATA_API_KEY3}`;
var EVENT_KEYUP_DATA_API = `keyup${EVENT_KEY4}${DATA_API_KEY3}`;
var CLASS_NAME_SHOW = "show";
var SELECTOR_DATA_TOGGLE3 = '[data-bs-toggle="menu"]:not(.disabled):not(:disabled)';
var SELECTOR_MENU = ".menu";
var SELECTOR_SUBMENU = ".submenu";
var SELECTOR_SUBMENU_TOGGLE = ".submenu > .menu-item";
var SELECTOR_NAVBAR_NAV = ".navbar-nav";
var SELECTOR_VISIBLE_ITEMS = ".menu-item:not(.disabled):not(:disabled)";
var DEFAULT_PLACEMENT = "bottom-start";
var SUBMENU_PLACEMENT = "end-start";
var resolveLogicalPlacement = (placement) => {
  if (isRTL()) return placement.replace(/^start(?=-|$)/, "right").replace(/^end(?=-|$)/, "left");
  return placement.replace(/^start(?=-|$)/, "left").replace(/^end(?=-|$)/, "right");
};
var triangleSign = (p1, p2, p3) => (p1.x - p3.x) * (p2.y - p3.y) - (p2.x - p3.x) * (p1.y - p3.y);
var Default4 = {
  autoClose: true,
  boundary: "clippingParents",
  container: false,
  display: "dynamic",
  offset: [0, 2],
  floatingConfig: null,
  menu: null,
  placement: DEFAULT_PLACEMENT,
  reference: "toggle",
  strategy: "absolute",
  submenuTrigger: "both",
  submenuDelay: SUBMENU_CLOSE_DELAY
};
var DefaultType4 = {
  autoClose: "(boolean|string)",
  boundary: "(string|element)",
  container: "(string|element|boolean)",
  display: "string",
  offset: "(array|string|function)",
  floatingConfig: "(null|object|function)",
  menu: "(null|element)",
  placement: "string",
  reference: "(string|element|object)",
  strategy: "string",
  submenuTrigger: "string",
  submenuDelay: "number"
};
var Menu = class Menu2 extends BaseComponent {
  static _openInstances = /* @__PURE__ */ new Set();
  constructor(element, config) {
    if (typeof computePosition2 === "undefined") throw new TypeError("Bootstrap's menus require Floating UI (https://floating-ui.com)");
    super(element, config);
    this._floatingCleanup = null;
    this._mediaQueryListeners = [];
    this._responsivePlacements = null;
    this._parent = this._element.parentNode;
    this._openSubmenus = /* @__PURE__ */ new Map();
    this._submenuCloseTimeouts = /* @__PURE__ */ new Map();
    this._hoverIntentData = null;
    this._menu = this._config.menu || this._findMenu();
    if (!this._config.menu && this._menu) this._parent = this._findWrapper(this._menu);
    this._isSubmenu = this._parent.classList?.contains("submenu");
    this._menuOriginalParent = this._menu?.parentNode;
    this._parseResponsivePlacements();
    this._setupSubmenuListeners();
    this._initSubmenuTriggers();
  }
  static get Default() {
    return Default4;
  }
  static get DefaultType() {
    return DefaultType4;
  }
  static get NAME() {
    return NAME4;
  }
  toggle() {
    return this._isShown() ? this.hide() : this.show();
  }
  async show() {
    if (isDisabled(this._element) || this._isShown()) return;
    const relatedTarget = { relatedTarget: this._element };
    if (EventHandler.trigger(this._element, EVENT_SHOW2, relatedTarget).defaultPrevented) return;
    this._moveMenuToContainer();
    this._createFloating();
    if ("ontouchstart" in document.documentElement && !this._parent.closest(SELECTOR_NAVBAR_NAV)) for (const element of document.body.children) EventHandler.on(element, "mouseover", noop);
    this._element.focus({ focusVisible: false });
    this._element.setAttribute("aria-expanded", "true");
    this._menu.classList.add(CLASS_NAME_SHOW);
    this._element.classList.add(CLASS_NAME_SHOW);
    if (this._parent) this._parent.classList.add(CLASS_NAME_SHOW);
    Menu2._openInstances.add(this);
    await this._queueCallback(() => {
      if (this._isShown()) EventHandler.trigger(this._element, EVENT_SHOWN, relatedTarget);
    }, this._menu, this._isAnimated());
  }
  async hide() {
    if (isDisabled(this._element) || !this._isShown()) return;
    const relatedTarget = { relatedTarget: this._element };
    await this._completeHide(relatedTarget);
  }
  dispose() {
    this._disposeFloating();
    this._restoreMenuToOriginalParent();
    this._disposeMediaQueryListeners();
    this._closeAllSubmenus();
    this._clearAllSubmenuTimeouts();
    Menu2._openInstances.delete(this);
    super.dispose();
  }
  update() {
    if (this._floatingCleanup) this._updateFloatingPosition();
  }
  _findMenu() {
    const wrapper = SelectorEngine.closest(this._element, `:has(${SELECTOR_MENU})`);
    return SelectorEngine.next(this._element, SELECTOR_MENU)[0] || SelectorEngine.prev(this._element, SELECTOR_MENU)[0] || SelectorEngine.findOne(SELECTOR_MENU, wrapper || this._parent);
  }
  _findWrapper(menu) {
    let wrapper = this._element.parentNode;
    while (wrapper instanceof Element && !wrapper.contains(menu)) wrapper = wrapper.parentNode;
    return wrapper instanceof Element ? wrapper : this._element.parentNode;
  }
  async _completeHide(relatedTarget) {
    if (EventHandler.trigger(this._element, EVENT_HIDE, relatedTarget).defaultPrevented) return;
    this._closeAllSubmenus();
    if ("ontouchstart" in document.documentElement) for (const element of document.body.children) EventHandler.off(element, "mouseover", noop);
    this._menu.classList.remove(CLASS_NAME_SHOW);
    this._element.classList.remove(CLASS_NAME_SHOW);
    if (this._parent) this._parent.classList.remove(CLASS_NAME_SHOW);
    this._element.setAttribute("aria-expanded", "false");
    Menu2._openInstances.delete(this);
    await this._queueCallback(() => {
      if (this._isShown()) return;
      this._disposeFloating();
      this._restoreMenuToOriginalParent();
      Manipulator.removeDataAttribute(this._menu, "placement");
      Manipulator.removeDataAttribute(this._menu, "display");
      EventHandler.trigger(this._element, EVENT_HIDDEN3, relatedTarget);
    }, this._menu, this._isAnimated());
  }
  _getConfig(config) {
    config = super._getConfig(config);
    if (typeof config.reference === "object" && !isElement(config.reference) && typeof config.reference.getBoundingClientRect !== "function") throw new TypeError(`${NAME4.toUpperCase()}: Option "reference" provided type "object" without a required "getBoundingClientRect" method.`);
    return config;
  }
  _createFloating() {
    if (this._config.display === "static") {
      Manipulator.setDataAttribute(this._menu, "display", "static");
      return;
    }
    let referenceElement = this._element;
    if (this._config.reference === "parent") referenceElement = this._parent;
    else if (isElement(this._config.reference)) referenceElement = getElement(this._config.reference);
    else if (typeof this._config.reference === "object") referenceElement = this._config.reference;
    this._updateFloatingPosition(referenceElement);
    this._floatingCleanup = autoUpdate(referenceElement, this._menu, () => this._updateFloatingPosition(referenceElement));
  }
  async _updateFloatingPosition(referenceElement = null) {
    if (!this._menu) return;
    if (!referenceElement) if (this._config.reference === "parent") referenceElement = this._parent;
    else if (isElement(this._config.reference)) referenceElement = getElement(this._config.reference);
    else if (typeof this._config.reference === "object") referenceElement = this._config.reference;
    else referenceElement = this._element;
    const placement = this._getPlacement();
    const middleware = this._getFloatingMiddleware();
    const floatingConfig = this._getFloatingConfig(placement, middleware);
    await this._applyFloatingPosition(referenceElement, this._menu, floatingConfig.placement, floatingConfig.middleware, floatingConfig.strategy);
  }
  _isShown() {
    return this._menu.classList.contains(CLASS_NAME_SHOW);
  }
  _isAnimated() {
    return getTransitionDurationFromElement(this._menu) > 0;
  }
  _getPlacement() {
    const placement = this._responsivePlacements ? getResponsivePlacement(this._responsivePlacements, DEFAULT_PLACEMENT) : this._config.placement;
    return resolveLogicalPlacement(placement);
  }
  _parseResponsivePlacements() {
    this._responsivePlacements = parseResponsivePlacement(this._config.placement, DEFAULT_PLACEMENT);
    if (this._responsivePlacements) this._setupMediaQueryListeners();
  }
  _setupMediaQueryListeners() {
    this._disposeMediaQueryListeners();
    this._mediaQueryListeners = createBreakpointListeners(() => {
      if (this._isShown()) this._updateFloatingPosition();
    });
  }
  _disposeMediaQueryListeners() {
    disposeBreakpointListeners(this._mediaQueryListeners);
    this._mediaQueryListeners = [];
  }
  _getOffset() {
    const { offset: offsetConfig } = this._config;
    if (typeof offsetConfig === "string") return offsetConfig.split(",").map((value) => Number.parseInt(value, 10));
    if (typeof offsetConfig === "function") return ({ placement, rects }) => {
      return toFloatingOffset(offsetConfig({
        placement,
        reference: rects.reference,
        floating: rects.floating
      }, this._element));
    };
    return offsetConfig;
  }
  _getFloatingMiddleware() {
    const offsetValue = this._getOffset();
    return [
      offset2(typeof offsetValue === "function" ? offsetValue : toFloatingOffset(offsetValue)),
      flip2({
        fallbackPlacements: this._getFallbackPlacements(),
        fallbackStrategy: "initialPlacement"
      }),
      shift2({ boundary: this._config.boundary === "clippingParents" ? "clippingAncestors" : this._config.boundary })
    ];
  }
  _getFallbackPlacements() {
    return {
      bottom: [
        "top",
        "bottom-start",
        "bottom-end",
        "top-start",
        "top-end"
      ],
      "bottom-start": [
        "top-start",
        "bottom-end",
        "top-end"
      ],
      "bottom-end": [
        "top-end",
        "bottom-start",
        "top-start"
      ],
      top: [
        "bottom",
        "top-start",
        "top-end",
        "bottom-start",
        "bottom-end"
      ],
      "top-start": [
        "bottom-start",
        "top-end",
        "bottom-end"
      ],
      "top-end": [
        "bottom-end",
        "top-start",
        "bottom-start"
      ],
      right: [
        "left",
        "right-start",
        "right-end",
        "left-start",
        "left-end"
      ],
      "right-start": [
        "left-start",
        "right-end",
        "left-end",
        "top-start",
        "bottom-start"
      ],
      "right-end": [
        "left-end",
        "right-start",
        "left-start",
        "top-end",
        "bottom-end"
      ],
      left: [
        "right",
        "left-start",
        "left-end",
        "right-start",
        "right-end"
      ],
      "left-start": [
        "right-start",
        "left-end",
        "right-end",
        "top-start",
        "bottom-start"
      ],
      "left-end": [
        "right-end",
        "left-start",
        "right-start",
        "top-end",
        "bottom-end"
      ]
    }[this._getPlacement()] || [
      "top",
      "bottom",
      "right",
      "left"
    ];
  }
  _getFloatingConfig(placement, middleware) {
    const defaultConfig = {
      placement,
      middleware,
      strategy: this._config.strategy
    };
    return {
      ...defaultConfig,
      ...execute(this._config.floatingConfig, [void 0, defaultConfig])
    };
  }
  _disposeFloating() {
    if (this._floatingCleanup) {
      this._floatingCleanup();
      this._floatingCleanup = null;
    }
  }
  _getContainer() {
    const { container } = this._config;
    if (container === false) return null;
    return container === true ? document.body : getElement(container);
  }
  _moveMenuToContainer() {
    const container = this._getContainer();
    if (!container || !this._menu) return;
    if (this._menu.parentNode !== container) container.append(this._menu);
  }
  _restoreMenuToOriginalParent() {
    if (!this._menuOriginalParent || !this._menu) return;
    if (this._menu.parentNode !== this._menuOriginalParent) this._menuOriginalParent.append(this._menu);
  }
  async _applyFloatingPosition(reference, floating, placement, middleware, strategy = "absolute") {
    if (!floating.isConnected) return null;
    const { x, y, placement: finalPlacement } = await computePosition2(reference, floating, {
      placement,
      middleware,
      strategy
    });
    if (!floating.isConnected) return null;
    Object.assign(floating.style, {
      position: strategy,
      left: `${x}px`,
      top: `${y}px`,
      margin: "0"
    });
    Manipulator.setDataAttribute(floating, "placement", finalPlacement);
    return finalPlacement;
  }
  _setupSubmenuListeners() {
    if (this._config.submenuTrigger === "hover" || this._config.submenuTrigger === "both") {
      EventHandler.on(this._menu, "mouseenter", SELECTOR_SUBMENU_TOGGLE, (event) => {
        this._onSubmenuTriggerEnter(event);
      });
      EventHandler.on(this._menu, "mouseleave", SELECTOR_SUBMENU, (event) => {
        this._onSubmenuLeave(event);
      });
      EventHandler.on(this._menu, "mousemove", (event) => {
        this._trackMousePosition(event);
      });
    }
    if (this._config.submenuTrigger === "click" || this._config.submenuTrigger === "both") EventHandler.on(this._menu, "click", SELECTOR_SUBMENU_TOGGLE, (event) => {
      this._onSubmenuTriggerClick(event);
    });
  }
  _onSubmenuTriggerEnter(event) {
    const trigger2 = event.target.closest(SELECTOR_SUBMENU_TOGGLE);
    if (!trigger2) return;
    const submenuWrapper = trigger2.closest(SELECTOR_SUBMENU);
    const submenu = SelectorEngine.findOne(SELECTOR_MENU, submenuWrapper);
    if (!submenu) return;
    this._cancelSubmenuCloseTimeout(submenu);
    this._closeSiblingSubmenus(submenuWrapper);
    this._openSubmenu(trigger2, submenu, submenuWrapper);
  }
  _onSubmenuLeave(event) {
    const submenuWrapper = event.target.closest(SELECTOR_SUBMENU);
    const submenu = SelectorEngine.findOne(SELECTOR_MENU, submenuWrapper);
    if (!submenu || !this._openSubmenus.has(submenu)) return;
    if (this._isMovingTowardSubmenu(event, submenu)) return;
    this._scheduleSubmenuClose(submenu, submenuWrapper);
  }
  _onSubmenuTriggerClick(event) {
    const trigger2 = event.target.closest(SELECTOR_SUBMENU_TOGGLE);
    if (!trigger2) return;
    event.preventDefault();
    event.stopPropagation();
    const submenuWrapper = trigger2.closest(SELECTOR_SUBMENU);
    const submenu = SelectorEngine.findOne(SELECTOR_MENU, submenuWrapper);
    if (!submenu) return;
    if (this._openSubmenus.has(submenu)) this._closeSubmenu(submenu, submenuWrapper);
    else {
      this._closeSiblingSubmenus(submenuWrapper);
      this._openSubmenu(trigger2, submenu, submenuWrapper);
    }
  }
  _initSubmenuTriggers() {
    if (!this._menu) return;
    for (const trigger2 of SelectorEngine.find(SELECTOR_SUBMENU_TOGGLE, this._menu)) {
      if (!SelectorEngine.findOne(SELECTOR_MENU, trigger2.parentElement)) continue;
      trigger2.setAttribute("aria-haspopup", "true");
      if (!trigger2.hasAttribute("aria-expanded")) trigger2.setAttribute("aria-expanded", "false");
    }
  }
  _openSubmenu(trigger2, submenu, submenuWrapper) {
    if (this._openSubmenus.has(submenu)) return;
    trigger2.setAttribute("aria-expanded", "true");
    trigger2.setAttribute("aria-haspopup", "true");
    submenu.style.opacity = "0";
    submenu.classList.add(CLASS_NAME_SHOW);
    submenuWrapper.classList.add(CLASS_NAME_SHOW);
    const cleanup = this._createSubmenuFloating(trigger2, submenu, submenuWrapper);
    this._openSubmenus.set(submenu, cleanup);
    EventHandler.on(submenu, "mouseenter", () => {
      this._cancelSubmenuCloseTimeout(submenu);
    });
  }
  _closeSubmenu(submenu, submenuWrapper) {
    if (!this._openSubmenus.has(submenu)) return;
    const nestedSubmenus = SelectorEngine.find(`${SELECTOR_SUBMENU} ${SELECTOR_MENU}.${CLASS_NAME_SHOW}`, submenu);
    for (const nested of nestedSubmenus) {
      const nestedWrapper = nested.closest(SELECTOR_SUBMENU);
      this._closeSubmenu(nested, nestedWrapper);
    }
    const trigger2 = SelectorEngine.findOne(SELECTOR_SUBMENU_TOGGLE, submenuWrapper);
    const cleanup = this._openSubmenus.get(submenu);
    if (cleanup) cleanup();
    this._openSubmenus.delete(submenu);
    EventHandler.off(submenu, "mouseenter");
    if (trigger2) trigger2.setAttribute("aria-expanded", "false");
    submenu.classList.remove(CLASS_NAME_SHOW);
    submenuWrapper.classList.remove(CLASS_NAME_SHOW);
    submenu.style.opacity = "";
  }
  _closeAllSubmenus() {
    for (const [submenu] of this._openSubmenus) {
      const submenuWrapper = submenu.closest(SELECTOR_SUBMENU);
      this._closeSubmenu(submenu, submenuWrapper);
    }
  }
  _closeSiblingSubmenus(currentSubmenuWrapper) {
    const parent = currentSubmenuWrapper.parentNode;
    const siblingSubmenus = SelectorEngine.find(`${SELECTOR_SUBMENU} > ${SELECTOR_MENU}.${CLASS_NAME_SHOW}`, parent);
    for (const siblingMenu of siblingSubmenus) {
      const siblingWrapper = siblingMenu.closest(SELECTOR_SUBMENU);
      if (siblingWrapper !== currentSubmenuWrapper) this._closeSubmenu(siblingMenu, siblingWrapper);
    }
  }
  _createSubmenuFloating(trigger2, submenu, submenuWrapper) {
    const referenceElement = submenuWrapper;
    const placement = resolveLogicalPlacement(SUBMENU_PLACEMENT);
    const middleware = [
      offset2({
        mainAxis: 0,
        crossAxis: -4
      }),
      flip2({ fallbackPlacements: [
        resolveLogicalPlacement("start-start"),
        resolveLogicalPlacement("end-end"),
        resolveLogicalPlacement("start-end")
      ] }),
      shift2({ padding: 8 })
    ];
    const updatePosition = () => this._applyFloatingPosition(referenceElement, submenu, placement, middleware).then((finalPlacement) => {
      submenu.style.opacity = "";
      return finalPlacement;
    });
    updatePosition();
    return autoUpdate(referenceElement, submenu, updatePosition);
  }
  _scheduleSubmenuClose(submenu, submenuWrapper) {
    this._cancelSubmenuCloseTimeout(submenu);
    const timeoutId = setTimeout(() => {
      this._closeSubmenu(submenu, submenuWrapper);
      this._submenuCloseTimeouts.delete(submenu);
    }, this._config.submenuDelay);
    this._submenuCloseTimeouts.set(submenu, timeoutId);
  }
  _cancelSubmenuCloseTimeout(submenu) {
    const timeoutId = this._submenuCloseTimeouts.get(submenu);
    if (timeoutId) {
      clearTimeout(timeoutId);
      this._submenuCloseTimeouts.delete(submenu);
    }
  }
  _clearAllSubmenuTimeouts() {
    for (const timeoutId of this._submenuCloseTimeouts.values()) clearTimeout(timeoutId);
    this._submenuCloseTimeouts.clear();
  }
  _trackMousePosition(event) {
    this._hoverIntentData = {
      x: event.clientX,
      y: event.clientY,
      timestamp: Date.now()
    };
  }
  _isMovingTowardSubmenu(event, submenu) {
    if (!this._hoverIntentData) return false;
    const submenuRect = submenu.getBoundingClientRect();
    const currentPos = {
      x: event.clientX,
      y: event.clientY
    };
    const lastPos = {
      x: this._hoverIntentData.x,
      y: this._hoverIntentData.y
    };
    const targetX = isRTL() ? submenuRect.right : submenuRect.left;
    const topCorner = {
      x: targetX,
      y: submenuRect.top
    };
    const bottomCorner = {
      x: targetX,
      y: submenuRect.bottom
    };
    return this._pointInTriangle(currentPos, lastPos, topCorner, bottomCorner);
  }
  _pointInTriangle(point, v1, v2, v3) {
    const d1 = triangleSign(point, v1, v2);
    const d2 = triangleSign(point, v2, v3);
    const d3 = triangleSign(point, v3, v1);
    return !((d1 < 0 || d2 < 0 || d3 < 0) && (d1 > 0 || d2 > 0 || d3 > 0));
  }
  _getItemsInMenu(menu) {
    return SelectorEngine.find(`:scope > ${SELECTOR_VISIBLE_ITEMS}, :scope > ${SELECTOR_SUBMENU} > ${SELECTOR_VISIBLE_ITEMS}`, menu).filter((element) => isVisible(element));
  }
  _selectMenuItem({ key, target }) {
    const currentMenu = target.closest(SELECTOR_MENU) || this._menu;
    const items = this._getItemsInMenu(currentMenu);
    if (!items.length) return;
    const nextItem = getNextActiveElement(items, target, key === ARROW_DOWN_KEY, !items.includes(target));
    nextItem.focus();
    this._closeUnrelatedSubmenus(currentMenu, nextItem);
  }
  _closeUnrelatedSubmenus(currentMenu, focusedElement) {
    for (const [submenu] of this._openSubmenus) {
      const submenuWrapper = submenu.closest(SELECTOR_SUBMENU);
      if (!submenuWrapper || submenuWrapper.parentElement !== currentMenu || submenuWrapper.contains(focusedElement)) continue;
      this._closeSubmenu(submenu, submenuWrapper);
    }
  }
  _handleSubmenuKeydown(event) {
    const { key, target } = event;
    const isRtl = isRTL();
    const enterKey = isRtl ? ARROW_LEFT_KEY : ARROW_RIGHT_KEY;
    const exitKey = isRtl ? ARROW_RIGHT_KEY : ARROW_LEFT_KEY;
    const submenuWrapper = target.closest(SELECTOR_SUBMENU);
    const isSubmenuToggle = target.matches(SELECTOR_SUBMENU_TOGGLE);
    if ((key === ENTER_KEY || key === SPACE_KEY || key === enterKey) && submenuWrapper && isSubmenuToggle) {
      event.preventDefault();
      event.stopPropagation();
      const submenu = SelectorEngine.findOne(SELECTOR_MENU, submenuWrapper);
      if (submenu) {
        this._closeSiblingSubmenus(submenuWrapper);
        this._openSubmenu(target, submenu, submenuWrapper);
        requestAnimationFrame(() => {
          const firstItem = SelectorEngine.findOne(SELECTOR_VISIBLE_ITEMS, submenu);
          if (firstItem) firstItem.focus();
        });
      }
      return true;
    }
    if (key === exitKey) {
      const currentMenu = target.closest(SELECTOR_MENU);
      const parentSubmenuWrapper = currentMenu?.closest(SELECTOR_SUBMENU);
      if (parentSubmenuWrapper) {
        event.preventDefault();
        event.stopPropagation();
        const parentTrigger = SelectorEngine.findOne(SELECTOR_SUBMENU_TOGGLE, parentSubmenuWrapper);
        this._closeSubmenu(currentMenu, parentSubmenuWrapper);
        if (parentTrigger) parentTrigger.focus();
        return true;
      }
    }
    if (key === HOME_KEY || key === END_KEY) {
      event.preventDefault();
      event.stopPropagation();
      const currentMenu = target.closest(SELECTOR_MENU);
      const items = this._getItemsInMenu(currentMenu);
      if (items.length) {
        const targetItem = key === HOME_KEY ? items[0] : items.at(-1);
        targetItem.focus();
        this._closeUnrelatedSubmenus(currentMenu, targetItem);
      }
      return true;
    }
    return false;
  }
  static clearMenus(event) {
    if (event.button === RIGHT_MOUSE_BUTTON || event.type === "keyup" && event.key !== TAB_KEY) return;
    for (const instance of Menu2._openInstances) {
      if (instance._config.autoClose === false) continue;
      const composedPath = event.composedPath();
      const isMenuTarget = composedPath.includes(instance._menu);
      if (composedPath.includes(instance._element) || instance._config.autoClose === "inside" && !isMenuTarget || instance._config.autoClose === "outside" && isMenuTarget) continue;
      const formAncestor = event.target.closest?.("form");
      const isInsideMenuForm = Boolean(formAncestor) && instance._menu.contains(formAncestor);
      if (instance._menu.contains(event.target) && (event.type === "keyup" && event.key === TAB_KEY || /input|select|option|textarea|form/i.test(event.target.tagName) || isInsideMenuForm)) continue;
      const relatedTarget = { relatedTarget: instance._element };
      if (event.type === "click") relatedTarget.clickEvent = event;
      instance._completeHide(relatedTarget);
    }
  }
  static _getToggleFromKeydownContext(element, event) {
    if (element.matches(SELECTOR_DATA_TOGGLE3)) return element;
    for (const instance of Menu2._openInstances) if (instance._element === event.target || instance._menu?.contains(event.target)) return instance._element;
    return SelectorEngine.prev(element, SELECTOR_DATA_TOGGLE3)[0] || SelectorEngine.next(element, SELECTOR_DATA_TOGGLE3)[0] || SelectorEngine.findOne(SELECTOR_DATA_TOGGLE3, event.delegateTarget.parentNode);
  }
  static dataApiKeydownHandler(event) {
    const isInput = /input|textarea/i.test(event.target.tagName) || event.target.isContentEditable;
    const isEscapeEvent = event.key === ESCAPE_KEY;
    const isUpOrDownEvent = [ARROW_UP_KEY, ARROW_DOWN_KEY].includes(event.key);
    const isLeftOrRightEvent = [ARROW_LEFT_KEY, ARROW_RIGHT_KEY].includes(event.key);
    const isHomeOrEndEvent = [HOME_KEY, END_KEY].includes(event.key);
    const isEnterOrSpaceEvent = [ENTER_KEY, SPACE_KEY].includes(event.key);
    const isSubmenuTrigger = event.target.matches(SELECTOR_SUBMENU_TOGGLE);
    if (!isUpOrDownEvent && !isEscapeEvent && !isLeftOrRightEvent && !isHomeOrEndEvent && !(isEnterOrSpaceEvent && isSubmenuTrigger)) return;
    if (isInput && !isEscapeEvent) return;
    const getToggleButton = Menu2._getToggleFromKeydownContext(this, event);
    if (!getToggleButton) return;
    const instance = Menu2.getOrCreateInstance(getToggleButton);
    if ((isLeftOrRightEvent || isHomeOrEndEvent || isEnterOrSpaceEvent && isSubmenuTrigger) && instance._handleSubmenuKeydown(event)) return;
    if (isUpOrDownEvent) {
      event.preventDefault();
      event.stopPropagation();
      instance.show();
      instance._selectMenuItem(event);
      return;
    }
    if (isEscapeEvent && instance._isShown()) {
      event.preventDefault();
      event.stopPropagation();
      const currentMenu = event.target.closest(SELECTOR_MENU);
      const parentSubmenuWrapper = currentMenu?.closest(SELECTOR_SUBMENU);
      if (parentSubmenuWrapper && instance._openSubmenus.size > 0) {
        const parentTrigger = SelectorEngine.findOne(SELECTOR_SUBMENU_TOGGLE, parentSubmenuWrapper);
        instance._closeSubmenu(currentMenu, parentSubmenuWrapper);
        if (parentTrigger) parentTrigger.focus();
        return;
      }
      instance.hide();
      getToggleButton.focus();
    }
  }
};
EventHandler.on(document, EVENT_KEYDOWN_DATA_API, SELECTOR_DATA_TOGGLE3, Menu.dataApiKeydownHandler);
EventHandler.on(document, EVENT_KEYDOWN_DATA_API, SELECTOR_MENU, Menu.dataApiKeydownHandler);
EventHandler.on(document, EVENT_CLICK_DATA_API3, Menu.clearMenus);
EventHandler.on(document, EVENT_KEYUP_DATA_API, Menu.clearMenus);
EventHandler.on(document, EVENT_CLICK_DATA_API3, SELECTOR_DATA_TOGGLE3, function(event) {
  event.preventDefault();
  Menu.getOrCreateInstance(this).toggle();
});

// node_modules/bootstrap/js/dist/combobox.js
var NAME5 = "combobox";
var EVENT_KEY5 = `.bs.combobox`;
var DATA_API_KEY4 = ".data-api";
var ESCAPE_KEY2 = "Escape";
var TAB_KEY2 = "Tab";
var ARROW_UP_KEY2 = "ArrowUp";
var ARROW_DOWN_KEY2 = "ArrowDown";
var HOME_KEY2 = "Home";
var END_KEY2 = "End";
var ENTER_KEY2 = "Enter";
var SPACE_KEY2 = " ";
var EVENT_CHANGE = `change${EVENT_KEY5}`;
var EVENT_SHOW3 = `show${EVENT_KEY5}`;
var EVENT_SHOWN2 = `shown${EVENT_KEY5}`;
var EVENT_HIDE2 = `hide${EVENT_KEY5}`;
var EVENT_HIDDEN4 = `hidden${EVENT_KEY5}`;
var EVENT_CLICK_DATA_API4 = `click${EVENT_KEY5}${DATA_API_KEY4}`;
var CLASS_NAME_SHOW2 = "show";
var CLASS_NAME_SELECTED = "selected";
var CLASS_NAME_PLACEHOLDER = "combobox-placeholder";
var SELECTOR_DATA_TOGGLE4 = '[data-bs-toggle="combobox"]';
var SELECTOR_MENU2 = ".menu";
var SELECTOR_MENU_ITEM = ".menu-item[data-bs-value]";
var SELECTOR_VISIBLE_ITEMS2 = ".menu-item[data-bs-value]:not(.disabled):not(:disabled)";
var SELECTOR_VALUE = ".combobox-value";
var SELECTOR_SEARCH_INPUT = ".combobox-search-input";
var SELECTOR_NO_RESULTS = ".combobox-no-results";
var Default5 = {
  boundary: "clippingParents",
  multiple: false,
  name: null,
  offset: [0, 2],
  placeholder: "",
  placement: "bottom-start",
  search: false,
  searchNormalize: false
};
var DefaultType5 = {
  boundary: "(string|element)",
  multiple: "boolean",
  name: "(string|null)",
  offset: "(array|string|function)",
  placeholder: "string",
  placement: "string",
  search: "boolean",
  searchNormalize: "boolean"
};
var Combobox = class extends BaseComponent {
  constructor(element, config) {
    super(element, config);
    this._toggle = this._element;
    this._menu = SelectorEngine.next(this._toggle, SELECTOR_MENU2)[0];
    this._valueDisplay = SelectorEngine.findOne(SELECTOR_VALUE, this._toggle);
    this._searchInput = SelectorEngine.findOne(SELECTOR_SEARCH_INPUT, this._menu);
    this._noResults = SelectorEngine.findOne(SELECTOR_NO_RESULTS, this._menu);
    this._hiddenInput = null;
    this._menuInstance = null;
    this._createHiddenInput();
    this._createMenuInstance();
    this._syncInitialSelection();
    this._addEventListeners();
  }
  static get Default() {
    return Default5;
  }
  static get DefaultType() {
    return DefaultType5;
  }
  static get NAME() {
    return NAME5;
  }
  toggle() {
    return this._isShown() ? this.hide() : this.show();
  }
  async show() {
    if (isDisabled(this._toggle) || this._isShown()) return;
    if (EventHandler.trigger(this._toggle, EVENT_SHOW3).defaultPrevented) return;
    this._menuInstance.show();
    if (this._searchInput) {
      this._searchInput.value = "";
      this._filterItems("");
      requestAnimationFrame(() => this._searchInput.focus());
    }
    EventHandler.trigger(this._toggle, EVENT_SHOWN2);
  }
  async hide() {
    if (!this._isShown()) return;
    if (EventHandler.trigger(this._toggle, EVENT_HIDE2).defaultPrevented) return;
    this._menuInstance.hide();
    EventHandler.trigger(this._toggle, EVENT_HIDDEN4);
  }
  dispose() {
    if (this._menuInstance) {
      this._menuInstance.dispose();
      this._menuInstance = null;
    }
    if (this._hiddenInput) {
      this._hiddenInput.remove();
      this._hiddenInput = null;
    }
    EventHandler.off(this._menu, EVENT_KEY5);
    EventHandler.off(this._toggle, EVENT_KEY5);
    EventHandler.off(this._searchInput, EVENT_KEY5);
    super.dispose();
  }
  _isShown() {
    return this._menu.classList.contains(CLASS_NAME_SHOW2);
  }
  _createHiddenInput() {
    const { name } = this._config;
    if (!name) return;
    this._hiddenInput = document.createElement("input");
    this._hiddenInput.type = "hidden";
    this._hiddenInput.name = name;
    this._hiddenInput.value = "";
    this._toggle.parentNode.insertBefore(this._hiddenInput, this._toggle);
  }
  _createMenuInstance() {
    this._menuInstance = new Menu(this._toggle, {
      menu: this._menu,
      autoClose: this._config.multiple ? "outside" : true,
      boundary: this._config.boundary,
      offset: this._config.offset,
      placement: this._config.placement
    });
  }
  _syncInitialSelection() {
    if (this._getSelectedItems().length > 0) {
      this._updateToggleText();
      this._updateHiddenInput();
    } else this._showPlaceholder();
  }
  _addEventListeners() {
    EventHandler.on(this._menu, `click${EVENT_KEY5}`, SELECTOR_MENU_ITEM, (event) => {
      const item = event.target.closest(SELECTOR_MENU_ITEM);
      if (!item || isDisabled(item)) return;
      event.preventDefault();
      event.stopPropagation();
      this._selectItem(item);
    });
    EventHandler.on(this._toggle, `keydown${EVENT_KEY5}`, (event) => {
      this._handleToggleKeydown(event);
    });
    EventHandler.on(this._menu, `keydown${EVENT_KEY5}`, (event) => {
      this._handleMenuKeydown(event);
    });
    if (this._searchInput) {
      EventHandler.on(this._searchInput, `input${EVENT_KEY5}`, () => {
        this._filterItems(this._searchInput.value);
      });
      EventHandler.on(this._searchInput, `keydown${EVENT_KEY5}`, (event) => {
        if (event.key === ARROW_DOWN_KEY2) {
          event.preventDefault();
          const items = this._getVisibleItems();
          if (items.length > 0) items[0].focus();
        }
        if (event.key === ESCAPE_KEY2) {
          this.hide();
          this._toggle.focus();
        }
      });
    }
  }
  _selectItem(item) {
    if (this._config.multiple) {
      item.classList.toggle(CLASS_NAME_SELECTED);
      setAriaAttribute(item, "aria-selected", item.classList.contains(CLASS_NAME_SELECTED));
    } else {
      const previouslySelected = SelectorEngine.find(`.${CLASS_NAME_SELECTED}`, this._menu);
      for (const prev of previouslySelected) {
        prev.classList.remove(CLASS_NAME_SELECTED);
        prev.setAttribute("aria-selected", "false");
      }
      item.classList.add(CLASS_NAME_SELECTED);
      item.setAttribute("aria-selected", "true");
    }
    this._updateToggleText();
    this._updateHiddenInput();
    const value = this._config.multiple ? this._getSelectedItems().map((el) => el.dataset.bsValue) : item.dataset.bsValue;
    EventHandler.trigger(this._toggle, EVENT_CHANGE, {
      value,
      item
    });
    if (!this._config.multiple) {
      this.hide();
      this._toggle.focus();
    }
  }
  _updateToggleText() {
    const selectedItems = this._getSelectedItems();
    if (selectedItems.length === 0) {
      this._showPlaceholder();
      return;
    }
    this._valueDisplay.classList.remove(CLASS_NAME_PLACEHOLDER);
    if (this._config.multiple && selectedItems.length > 1) this._valueDisplay.textContent = `${selectedItems.length} selected`;
    else {
      const item = selectedItems[0];
      const label = SelectorEngine.findOne(".menu-item-content > span:first-child", item);
      this._valueDisplay.textContent = label ? label.textContent : item.textContent.trim();
    }
  }
  _showPlaceholder() {
    const { placeholder } = this._config;
    if (placeholder) {
      this._valueDisplay.textContent = placeholder;
      this._valueDisplay.classList.add(CLASS_NAME_PLACEHOLDER);
    }
  }
  _updateHiddenInput() {
    if (!this._hiddenInput) return;
    const values = this._getSelectedItems().map((el) => el.dataset.bsValue);
    this._hiddenInput.value = this._config.multiple ? values.join(",") : values[0] || "";
  }
  _getSelectedItems() {
    return SelectorEngine.find(`.${CLASS_NAME_SELECTED}`, this._menu);
  }
  _getVisibleItems() {
    return SelectorEngine.find(SELECTOR_VISIBLE_ITEMS2, this._menu).filter((item) => isVisible(item));
  }
  _filterItems(query) {
    const normalizedQuery = this._normalizeText(query.toLowerCase().trim());
    const items = SelectorEngine.find(SELECTOR_MENU_ITEM, this._menu);
    let visibleCount = 0;
    for (const item of items) {
      const text = this._normalizeText(item.textContent.toLowerCase().trim());
      const matches = !normalizedQuery || text.includes(normalizedQuery);
      item.style.display = matches ? "" : "none";
      if (matches) visibleCount++;
    }
    if (this._noResults) this._noResults.classList.toggle("d-none", visibleCount > 0);
  }
  _normalizeText(text) {
    if (this._config.searchNormalize) return text.normalize("NFD").replace(/[\u0300-\u036F]/g, "");
    return text;
  }
  _handleToggleKeydown(event) {
    const { key } = event;
    if (key === ARROW_DOWN_KEY2 || key === ARROW_UP_KEY2) {
      event.preventDefault();
      if (!this._isShown()) this.show();
      const items = this._getVisibleItems();
      if (items.length > 0) (key === ARROW_DOWN_KEY2 ? items[0] : items.at(-1)).focus();
      return;
    }
    if ((key === ENTER_KEY2 || key === SPACE_KEY2) && !this._isShown()) {
      event.preventDefault();
      this.show();
    }
  }
  _handleMenuKeydown(event) {
    const { key, target } = event;
    if (key === ESCAPE_KEY2) {
      event.preventDefault();
      event.stopPropagation();
      this.hide();
      this._toggle.focus();
      return;
    }
    if (key === TAB_KEY2) {
      this.hide();
      return;
    }
    const isInput = target.matches("input");
    if (key === ARROW_DOWN_KEY2 || key === ARROW_UP_KEY2) {
      event.preventDefault();
      const items = this._getVisibleItems();
      if (items.length > 0) getNextActiveElement(items, target, key === ARROW_DOWN_KEY2, !items.includes(target)).focus();
      return;
    }
    if (key === HOME_KEY2 || key === END_KEY2) {
      event.preventDefault();
      const items = this._getVisibleItems();
      if (items.length > 0) (key === HOME_KEY2 ? items[0] : items.at(-1)).focus();
      return;
    }
    if ((key === ENTER_KEY2 || key === SPACE_KEY2) && !isInput) {
      event.preventDefault();
      const item = target.closest(SELECTOR_MENU_ITEM);
      if (item && !isDisabled(item)) this._selectItem(item);
    }
  }
};
EventHandler.on(document, EVENT_CLICK_DATA_API4, SELECTOR_DATA_TOGGLE4, function(event) {
  event.preventDefault();
  Combobox.getOrCreateInstance(this).toggle();
});
EventHandler.on(document, "DOMContentLoaded", () => {
  for (const toggle of SelectorEngine.find(SELECTOR_DATA_TOGGLE4)) Combobox.getOrCreateInstance(toggle);
});

// node_modules/bootstrap/js/dist/util/sanitizer.js
var ARIA_ATTRIBUTE_PATTERN = /^aria-[\w-]*$/i;
var DefaultAllowlist = {
  "*": [
    "class",
    "dir",
    "id",
    "lang",
    "role",
    ARIA_ATTRIBUTE_PATTERN
  ],
  a: [
    "target",
    "href",
    "title",
    "rel"
  ],
  area: [],
  b: [],
  br: [],
  col: [],
  code: [],
  dd: [],
  div: [],
  dl: [],
  dt: [],
  em: [],
  hr: [],
  h1: [],
  h2: [],
  h3: [],
  h4: [],
  h5: [],
  h6: [],
  i: [],
  img: [
    "src",
    "srcset",
    "alt",
    "title",
    "width",
    "height"
  ],
  li: [],
  ol: [],
  p: [],
  pre: [],
  s: [],
  small: [],
  span: [],
  sub: [],
  sup: [],
  strong: [],
  u: [],
  ul: []
};
var DefaultIconAllowlist = {
  "*": [
    "class",
    "role",
    ARIA_ATTRIBUTE_PATTERN
  ],
  svg: [
    "xmlns",
    "width",
    "height",
    "viewbox",
    "fill",
    "stroke",
    "stroke-width",
    "stroke-linecap",
    "stroke-linejoin",
    "focusable"
  ],
  path: [
    "d",
    "fill",
    "stroke",
    "stroke-width",
    "fill-rule",
    "clip-rule"
  ],
  line: [
    "x1",
    "y1",
    "x2",
    "y2",
    "stroke",
    "stroke-width",
    "stroke-linecap"
  ],
  circle: [
    "cx",
    "cy",
    "r",
    "fill",
    "stroke",
    "stroke-width"
  ],
  rect: [
    "x",
    "y",
    "width",
    "height",
    "rx",
    "ry",
    "fill",
    "stroke",
    "stroke-width"
  ],
  polyline: [
    "points",
    "fill",
    "stroke",
    "stroke-width"
  ],
  polygon: [
    "points",
    "fill",
    "stroke",
    "stroke-width"
  ],
  g: [
    "fill",
    "stroke",
    "stroke-width",
    "transform"
  ],
  span: [],
  i: []
};
var uriAttributes = /* @__PURE__ */ new Set([
  "background",
  "cite",
  "href",
  "itemtype",
  "longdesc",
  "poster",
  "src",
  "xlink:href"
]);
var SAFE_URL_PATTERN = /^(?!(?:javascript|data|vbscript):)(?:[a-z0-9+.-]+:|[^&:/?#]*(?:[/?#]|$))/i;
var DATA_URL_PATTERN = /^data:(?:image\/(?:bmp|gif|jpeg|jpg|png|tiff|webp)|video\/(?:mpeg|mp4|ogg|webm)|audio\/(?:mp3|oga|ogg|opus));base64,[\d+/a-z=]+$/i;
var allowedAttribute = (attribute, allowedAttributeList) => {
  const attributeName = attribute.nodeName.toLowerCase();
  if (allowedAttributeList.includes(attributeName)) {
    if (uriAttributes.has(attributeName)) return Boolean(SAFE_URL_PATTERN.test(attribute.nodeValue) || DATA_URL_PATTERN.test(attribute.nodeValue));
    return true;
  }
  return allowedAttributeList.filter((attributeRegex) => attributeRegex instanceof RegExp).some((regex) => regex.test(attributeName));
};
function sanitizeHtml(unsafeHtml, allowList, sanitizeFunction) {
  if (!unsafeHtml.length) return unsafeHtml;
  if (sanitizeFunction && typeof sanitizeFunction === "function") return sanitizeFunction(unsafeHtml);
  const createdDocument = new window.DOMParser().parseFromString(unsafeHtml, "text/html");
  const elements = [...createdDocument.body.querySelectorAll("*")];
  for (const element of elements) {
    const elementName = element.nodeName.toLowerCase();
    if (!Object.keys(allowList).includes(elementName)) {
      element.remove();
      continue;
    }
    const attributeList = [...element.attributes];
    const allowedAttributes = [...allowList["*"] || [], ...allowList[elementName] || []];
    for (const attribute of attributeList) if (!allowedAttribute(attribute, allowedAttributes)) element.removeAttribute(attribute.nodeName);
  }
  return createdDocument.body.innerHTML;
}

// node_modules/bootstrap/js/dist/chips.js
var NAME6 = "chips";
var EVENT_KEY6 = `.bs.chips`;
var DATA_API_KEY5 = ".data-api";
var EVENT_ADD = `add${EVENT_KEY6}`;
var EVENT_REMOVE = `remove${EVENT_KEY6}`;
var EVENT_CHANGE2 = `change${EVENT_KEY6}`;
var EVENT_SELECT = `select${EVENT_KEY6}`;
var SELECTOR_DATA_CHIPS = "[data-bs-chips]";
var SELECTOR_GHOST_INPUT = ".form-ghost";
var SELECTOR_CHIP = ".chip";
var SELECTOR_CHIP_DISMISS = ".chip-dismiss";
var CLASS_NAME_CHIP = "chip";
var CLASS_NAME_CHIP_DISMISS = "chip-dismiss";
var CLASS_NAME_ACTIVE = "active";
var Default6 = {
  separator: ",",
  allowDuplicates: false,
  maxChips: null,
  placeholder: "",
  dismissible: true,
  dismissIcon: '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="4" y1="4" x2="12" y2="12"/><line x1="12" y1="4" x2="4" y2="12"/></svg>',
  createOnBlur: true
};
var DefaultType6 = {
  separator: "(string|null)",
  allowDuplicates: "boolean",
  maxChips: "(number|null)",
  placeholder: "string",
  dismissible: "boolean",
  dismissIcon: "string",
  createOnBlur: "boolean"
};
var Chips = class extends BaseComponent {
  constructor(element, config) {
    super(element, config);
    const ghostInput = SelectorEngine.findOne(SELECTOR_GHOST_INPUT, this._element);
    this._chips = [];
    this._selectedChips = /* @__PURE__ */ new Set();
    this._anchorChip = null;
    if (ghostInput) this._input = ghostInput;
    else this._createInput();
    this._initializeExistingChips();
    this._addEventListeners();
  }
  static get Default() {
    return Default6;
  }
  static get DefaultType() {
    return DefaultType6;
  }
  static get NAME() {
    return NAME6;
  }
  add(value) {
    const trimmedValue = String(value).trim();
    if (!trimmedValue) return null;
    if (!this._config.allowDuplicates && this._chips.includes(trimmedValue)) return null;
    if (this._config.maxChips !== null && this._chips.length >= this._config.maxChips) return null;
    if (EventHandler.trigger(this._element, EVENT_ADD, {
      value: trimmedValue,
      relatedTarget: this._input
    }).defaultPrevented) return null;
    const chip = this._createChip(trimmedValue);
    this._element.insertBefore(chip, this._input);
    this._chips.push(trimmedValue);
    EventHandler.trigger(this._element, EVENT_CHANGE2, { values: this.getValues() });
    return chip;
  }
  remove(chipOrValue) {
    let chip;
    let value;
    if (typeof chipOrValue === "string") {
      value = chipOrValue;
      chip = this._findChipByValue(value);
    } else {
      chip = chipOrValue;
      value = this._getChipValue(chip);
    }
    if (!chip || !value) return false;
    if (EventHandler.trigger(this._element, EVENT_REMOVE, {
      value,
      chip,
      relatedTarget: this._input
    }).defaultPrevented) return false;
    this._selectedChips.delete(chip);
    if (this._anchorChip === chip) this._anchorChip = null;
    chip.remove();
    this._chips = this._getChipElements().map((chipElement) => this._getChipValue(chipElement)).filter(Boolean);
    EventHandler.trigger(this._element, EVENT_CHANGE2, { values: this.getValues() });
    return true;
  }
  removeSelected() {
    const chipsToRemove = [...this._selectedChips];
    for (const chip of chipsToRemove) this.remove(chip);
    this._input?.focus();
  }
  getValues() {
    return [...this._chips];
  }
  getSelectedValues() {
    return [...this._selectedChips].map((chip) => this._getChipValue(chip));
  }
  clear() {
    const chips = SelectorEngine.find(SELECTOR_CHIP, this._element);
    for (const chip of chips) chip.remove();
    this._chips = [];
    this._selectedChips.clear();
    this._anchorChip = null;
    EventHandler.trigger(this._element, EVENT_CHANGE2, { values: [] });
  }
  clearSelection() {
    for (const chip of this._selectedChips) chip.classList.remove(CLASS_NAME_ACTIVE);
    this._selectedChips.clear();
    this._anchorChip = null;
    EventHandler.trigger(this._element, EVENT_SELECT, { selected: [] });
  }
  selectChip(chip, options = {}) {
    const { addToSelection = false, rangeSelect = false } = options;
    const chipElements = this._getChipElements();
    if (!chipElements.includes(chip)) return;
    if (rangeSelect && this._anchorChip) {
      const anchorIndex = chipElements.indexOf(this._anchorChip);
      const chipIndex = chipElements.indexOf(chip);
      const start = Math.min(anchorIndex, chipIndex);
      const end = Math.max(anchorIndex, chipIndex);
      if (!addToSelection) {
        const anchorChip = this._anchorChip;
        this.clearSelection();
        this._anchorChip = anchorChip;
      }
      for (let i = start; i <= end; i++) {
        this._selectedChips.add(chipElements[i]);
        chipElements[i].classList.add(CLASS_NAME_ACTIVE);
      }
    } else if (addToSelection) if (this._selectedChips.has(chip)) {
      this._selectedChips.delete(chip);
      chip.classList.remove(CLASS_NAME_ACTIVE);
    } else {
      this._selectedChips.add(chip);
      chip.classList.add(CLASS_NAME_ACTIVE);
      this._anchorChip = chip;
    }
    else {
      this.clearSelection();
      this._selectedChips.add(chip);
      chip.classList.add(CLASS_NAME_ACTIVE);
      this._anchorChip = chip;
    }
    EventHandler.trigger(this._element, EVENT_SELECT, { selected: this.getSelectedValues() });
  }
  focus() {
    this._input?.focus();
  }
  dispose() {
    EventHandler.off(this._input, EVENT_KEY6);
    super.dispose();
  }
  _getChipElements() {
    return SelectorEngine.find(SELECTOR_CHIP, this._element);
  }
  _createInput() {
    const input = document.createElement("input");
    input.type = "text";
    input.className = "form-ghost";
    if (this._config.placeholder) input.placeholder = this._config.placeholder;
    this._element.append(input);
    this._input = input;
  }
  _initializeExistingChips() {
    const existingChips = SelectorEngine.find(SELECTOR_CHIP, this._element);
    for (const chip of existingChips) {
      const value = this._getChipValue(chip);
      if (value) {
        this._chips.push(value);
        this._setupChip(chip);
      }
    }
  }
  _setupChip(chip) {
    chip.setAttribute("tabindex", "0");
    if (this._config.dismissible && !SelectorEngine.findOne(SELECTOR_CHIP_DISMISS, chip)) chip.append(this._createDismissButton());
  }
  _createChip(value) {
    const chip = document.createElement("span");
    chip.className = CLASS_NAME_CHIP;
    chip.dataset.bsChipValue = value;
    chip.append(document.createTextNode(value));
    this._setupChip(chip);
    return chip;
  }
  _createDismissButton() {
    const button = document.createElement("button");
    button.type = "button";
    button.className = CLASS_NAME_CHIP_DISMISS;
    button.setAttribute("aria-label", "Remove");
    button.setAttribute("tabindex", "-1");
    button.innerHTML = sanitizeHtml(this._config.dismissIcon, DefaultIconAllowlist);
    return button;
  }
  _findChipByValue(value) {
    return this._getChipElements().find((chip) => this._getChipValue(chip) === value);
  }
  _getChipValue(chip) {
    if (chip.dataset.bsChipValue) return chip.dataset.bsChipValue;
    const clone = chip.cloneNode(true);
    const dismiss = SelectorEngine.findOne(SELECTOR_CHIP_DISMISS, clone);
    if (dismiss) dismiss.remove();
    return clone.textContent?.trim() || "";
  }
  _addEventListeners() {
    EventHandler.on(this._input, `keydown${EVENT_KEY6}`, (event) => this._handleInputKeydown(event));
    EventHandler.on(this._input, `input${EVENT_KEY6}`, (event) => this._handleInput(event));
    EventHandler.on(this._input, `paste${EVENT_KEY6}`, (event) => this._handlePaste(event));
    EventHandler.on(this._input, `focus${EVENT_KEY6}`, () => this.clearSelection());
    if (this._config.createOnBlur) EventHandler.on(this._input, `blur${EVENT_KEY6}`, (event) => {
      if (!event.relatedTarget?.closest(SELECTOR_CHIP)) this._createChipFromInput();
    });
    EventHandler.on(this._element, `click${EVENT_KEY6}`, SELECTOR_CHIP, (event) => {
      if (event.target.closest(SELECTOR_CHIP_DISMISS)) return;
      const chip = event.target.closest(SELECTOR_CHIP);
      if (chip) {
        event.preventDefault();
        this.selectChip(chip, {
          addToSelection: event.metaKey || event.ctrlKey,
          rangeSelect: event.shiftKey
        });
        chip.focus();
      }
    });
    EventHandler.on(this._element, `click${EVENT_KEY6}`, SELECTOR_CHIP_DISMISS, (event) => {
      event.stopPropagation();
      const chip = event.target.closest(SELECTOR_CHIP);
      if (chip) {
        this.remove(chip);
        this._input?.focus();
      }
    });
    EventHandler.on(this._element, `keydown${EVENT_KEY6}`, SELECTOR_CHIP, (event) => {
      this._handleChipKeydown(event);
    });
    EventHandler.on(this._element, `click${EVENT_KEY6}`, (event) => {
      if (event.target === this._element) {
        this.clearSelection();
        this._input?.focus();
      }
    });
  }
  _handleInputKeydown(event) {
    const { key } = event;
    switch (key) {
      case "Enter":
        event.preventDefault();
        this._createChipFromInput();
        break;
      case "Backspace":
      case "Delete":
        if (this._input.value === "") {
          event.preventDefault();
          const chips = this._getChipElements();
          if (chips.length > 0) {
            const lastChip = chips.at(-1);
            this.selectChip(lastChip);
            lastChip.focus();
          }
        }
        break;
      case "ArrowLeft":
        if (this._input.selectionStart === 0 && this._input.selectionEnd === 0) {
          event.preventDefault();
          const chips = this._getChipElements();
          if (chips.length > 0) {
            const lastChip = chips.at(-1);
            if (event.shiftKey) this.selectChip(lastChip, { addToSelection: true });
            else this.selectChip(lastChip);
            lastChip.focus();
          }
        }
        break;
      case "Escape":
        this._input.value = "";
        this.clearSelection();
        this._input.blur();
        break;
    }
  }
  _handleChipKeydown(event) {
    const { key } = event;
    const chip = event.target.closest(SELECTOR_CHIP);
    if (!chip) return;
    const chips = this._getChipElements();
    const currentIndex = chips.indexOf(chip);
    switch (key) {
      case "Backspace":
      case "Delete":
        event.preventDefault();
        this._handleChipDelete(currentIndex, chips);
        break;
      case "ArrowLeft":
        event.preventDefault();
        this._navigateChip(chips, currentIndex, -1, event.shiftKey);
        break;
      case "ArrowRight":
        event.preventDefault();
        this._navigateChip(chips, currentIndex, 1, event.shiftKey);
        break;
      case "Home":
        event.preventDefault();
        this._navigateToEdge(chips, 0, event.shiftKey);
        break;
      case "End":
        event.preventDefault();
        this.clearSelection();
        this._input?.focus();
        break;
      case "a":
        this._handleSelectAll(event, chips);
        break;
      case "Escape":
        event.preventDefault();
        this.clearSelection();
        this._input?.focus();
        break;
    }
  }
  _handleChipDelete(currentIndex, chips) {
    if (this._selectedChips.size === 0) return;
    const nextIndex = Math.min(currentIndex, chips.length - this._selectedChips.size - 1);
    this.removeSelected();
    const remainingChips = this._getChipElements();
    if (remainingChips.length > 0) {
      const focusIndex = Math.max(0, Math.min(nextIndex, remainingChips.length - 1));
      remainingChips[focusIndex].focus();
      this.selectChip(remainingChips[focusIndex]);
    } else this._input?.focus();
  }
  _navigateChip(chips, currentIndex, direction, shiftKey) {
    const targetIndex = currentIndex + direction;
    if (direction < 0 && targetIndex >= 0) {
      const targetChip = chips[targetIndex];
      this.selectChip(targetChip, shiftKey ? {
        addToSelection: true,
        rangeSelect: true
      } : {});
      targetChip.focus();
    } else if (direction > 0 && targetIndex < chips.length) {
      const targetChip = chips[targetIndex];
      this.selectChip(targetChip, shiftKey ? {
        addToSelection: true,
        rangeSelect: true
      } : {});
      targetChip.focus();
    } else if (direction > 0) {
      this.clearSelection();
      this._input?.focus();
    }
  }
  _navigateToEdge(chips, targetIndex, shiftKey) {
    if (chips.length === 0) return;
    const targetChip = chips[targetIndex];
    this.selectChip(targetChip, shiftKey ? { rangeSelect: true } : {});
    targetChip.focus();
  }
  _handleSelectAll(event, chips) {
    if (!(event.metaKey || event.ctrlKey)) return;
    event.preventDefault();
    for (const c of chips) {
      this._selectedChips.add(c);
      c.classList.add(CLASS_NAME_ACTIVE);
    }
    EventHandler.trigger(this._element, EVENT_SELECT, { selected: this.getSelectedValues() });
  }
  _handleInput(event) {
    const { value } = event.target;
    const { separator } = this._config;
    if (separator && value.includes(separator)) {
      const parts = value.split(separator);
      for (const part of parts.slice(0, -1)) this.add(part.trim());
      this._input.value = parts.at(-1);
    }
  }
  _handlePaste(event) {
    const { separator } = this._config;
    if (!separator) return;
    const pastedData = (event.clipboardData || window.clipboardData).getData("text");
    if (pastedData.includes(separator)) {
      event.preventDefault();
      const parts = pastedData.split(separator);
      for (const part of parts) this.add(part.trim());
    }
  }
  _createChipFromInput() {
    const value = this._input.value.trim();
    if (value) {
      this.add(value);
      this._input.value = "";
    }
  }
};
EventHandler.on(document, `DOMContentLoaded${EVENT_KEY6}${DATA_API_KEY5}`, () => {
  for (const element of SelectorEngine.find(SELECTOR_DATA_CHIPS)) Chips.getOrCreateInstance(element);
});

// node_modules/bootstrap/js/dist/util/template-factory.js
var NAME7 = "TemplateFactory";
var Default7 = {
  allowList: DefaultAllowlist,
  content: {},
  extraClass: "",
  html: false,
  sanitize: true,
  sanitizeFn: null,
  template: "<div></div>"
};
var DefaultType7 = {
  allowList: "object",
  content: "object",
  extraClass: "(string|function)",
  html: "boolean",
  sanitize: "boolean",
  sanitizeFn: "(null|function)",
  template: "string"
};
var DefaultContentType = {
  entry: "(string|element|function|null)",
  selector: "(string|element)"
};
var TemplateFactory = class extends Config {
  constructor(config) {
    super();
    this._config = this._getConfig(config);
  }
  static get Default() {
    return Default7;
  }
  static get DefaultType() {
    return DefaultType7;
  }
  static get NAME() {
    return NAME7;
  }
  getContent() {
    return Object.values(this._config.content).map((config) => this._resolvePossibleFunction(config)).filter(Boolean);
  }
  hasContent() {
    return this.getContent().length > 0;
  }
  changeContent(content) {
    this._checkContent(content);
    this._config.content = {
      ...this._config.content,
      ...content
    };
    return this;
  }
  toHtml() {
    const templateWrapper = document.createElement("div");
    templateWrapper.innerHTML = this._maybeSanitize(this._config.template);
    for (const [selector, text] of Object.entries(this._config.content)) this._setContent(templateWrapper, text, selector);
    const template = templateWrapper.children[0];
    const extraClass = this._resolvePossibleFunction(this._config.extraClass);
    if (extraClass) template.classList.add(...extraClass.split(" "));
    return template;
  }
  _typeCheckConfig(config) {
    super._typeCheckConfig(config);
    this._checkContent(config.content);
  }
  _checkContent(arg) {
    for (const [selector, content] of Object.entries(arg)) super._typeCheckConfig({
      selector,
      entry: content
    }, DefaultContentType);
  }
  _setContent(template, content, selector) {
    const templateElement = SelectorEngine.findOne(selector, template);
    if (!templateElement) return;
    content = this._resolvePossibleFunction(content);
    if (!content) {
      templateElement.remove();
      return;
    }
    if (isElement(content)) {
      this._putElementInTemplate(getElement(content), templateElement);
      return;
    }
    if (this._config.html) {
      templateElement.innerHTML = this._maybeSanitize(content);
      return;
    }
    templateElement.textContent = content;
  }
  _maybeSanitize(arg) {
    return this._config.sanitize ? sanitizeHtml(arg, this._config.allowList, this._config.sanitizeFn) : arg;
  }
  _resolvePossibleFunction(arg) {
    return execute(arg, [void 0, this]);
  }
  _putElementInTemplate(element, templateElement) {
    if (this._config.html) {
      templateElement.innerHTML = "";
      templateElement.append(element);
      return;
    }
    templateElement.textContent = element.textContent;
  }
};

// node_modules/bootstrap/js/dist/tooltip.js
var NAME8 = "tooltip";
var DISALLOWED_ATTRIBUTES = /* @__PURE__ */ new Set([
  "sanitize",
  "allowList",
  "sanitizeFn"
]);
var ESCAPE_KEY3 = "Escape";
var CLASS_NAME_MODAL = "modal";
var CLASS_NAME_SHOW3 = "show";
var SELECTOR_TOOLTIP_INNER = ".tooltip-inner";
var SELECTOR_MODAL = `.${CLASS_NAME_MODAL}`;
var SELECTOR_DATA_TOGGLE5 = '[data-bs-toggle="tooltip"]';
var EVENT_MODAL_HIDE = "hide.bs.modal";
var TRIGGER_HOVER = "hover";
var TRIGGER_FOCUS = "focus";
var TRIGGER_CLICK = "click";
var TRIGGER_MANUAL = "manual";
var EVENT_HIDE3 = "hide";
var EVENT_HIDDEN5 = "hidden";
var EVENT_SHOW4 = "show";
var EVENT_SHOWN3 = "shown";
var EVENT_INSERTED = "inserted";
var EVENT_CLICK = "click";
var EVENT_FOCUSIN = "focusin";
var EVENT_FOCUSOUT = "focusout";
var EVENT_MOUSEENTER = "mouseenter";
var EVENT_MOUSELEAVE = "mouseleave";
var EVENT_POINTERDOWN2 = "pointerdown";
var EVENT_POINTERUP2 = "pointerup";
var EVENT_KEYDOWN = "keydown";
var AttachmentMap = {
  AUTO: "auto",
  TOP: "top",
  RIGHT: isRTL() ? "left" : "right",
  BOTTOM: "bottom",
  LEFT: isRTL() ? "right" : "left"
};
var Default8 = {
  allowList: DefaultAllowlist,
  animation: true,
  boundary: "clippingParents",
  container: false,
  customClass: "",
  delay: 0,
  fallbackPlacements: [
    "top",
    "right",
    "bottom",
    "left"
  ],
  html: false,
  offset: [0, 6],
  placement: "top",
  floatingConfig: null,
  sanitize: true,
  sanitizeFn: null,
  selector: false,
  template: '<div class="tooltip" role="tooltip"><div class="tooltip-arrow"></div><div class="tooltip-inner"></div></div>',
  title: "",
  trigger: "hover focus"
};
var DefaultType8 = {
  allowList: "object",
  animation: "boolean",
  boundary: "(string|element)",
  container: "(string|element|boolean)",
  customClass: "(string|function)",
  delay: "(number|object)",
  fallbackPlacements: "array",
  html: "boolean",
  offset: "(array|string|function)",
  placement: "(string|function)",
  floatingConfig: "(null|object|function)",
  sanitize: "boolean",
  sanitizeFn: "(null|function)",
  selector: "(string|boolean)",
  template: "string",
  title: "(string|element|function)",
  trigger: "string"
};
var Tooltip = class extends BaseComponent {
  constructor(element, config) {
    if (typeof computePosition2 === "undefined") throw new TypeError("Bootstrap's tooltips require Floating UI (https://floating-ui.com)");
    super(element, config);
    this._isEnabled = true;
    this._timeout = 0;
    this._resolveTimeout = null;
    this._isHovered = null;
    this._activeTrigger = {};
    this._floatingCleanup = null;
    this._keydownHandler = null;
    this._tipEventOut = null;
    this._outsidePointerHandler = null;
    this._tipPointerUpHandler = null;
    this._tipPointerDown = false;
    this._templateFactory = null;
    this._newContent = null;
    this._mediaQueryListeners = [];
    this._responsivePlacements = null;
    this.tip = null;
    this._parseResponsivePlacements();
    this._setListeners();
    if (!this._config.selector) this._fixTitle();
  }
  static get Default() {
    return Default8;
  }
  static get DefaultType() {
    return DefaultType8;
  }
  static get NAME() {
    return NAME8;
  }
  enable() {
    this._isEnabled = true;
  }
  disable() {
    this._isEnabled = false;
  }
  toggleEnabled() {
    this._isEnabled = !this._isEnabled;
  }
  toggle() {
    if (!this._isEnabled) return Promise.resolve();
    return this._isShown() ? this._leave() : this._enter();
  }
  dispose() {
    this._clearTimeout();
    this._removeEscapeListener();
    this._removeFocusOutsideListener();
    EventHandler.off(this._element.closest(SELECTOR_MODAL), EVENT_MODAL_HIDE, this._hideModalHandler);
    if (this._element.getAttribute("data-bs-original-title")) this._element.setAttribute("title", this._element.getAttribute("data-bs-original-title"));
    this._disposeFloating();
    this._disposeMediaQueryListeners();
    super.dispose();
  }
  async show() {
    if (this._element.style.display === "none") throw new Error("Please use show on visible elements");
    if (!(this._isWithContent() && this._isEnabled)) return;
    const showEvent = EventHandler.trigger(this._element, this.constructor.eventName(EVENT_SHOW4));
    const isInTheDom = (findShadowRoot(this._element) || this._element.ownerDocument.documentElement).contains(this._element);
    if (showEvent.defaultPrevented || !isInTheDom) {
      this._isHovered = false;
      return;
    }
    this._disposeFloating();
    const tip = this._getTipElement();
    this._element.setAttribute("aria-describedby", tip.getAttribute("id"));
    let { container } = this._config;
    const closestDialog = this._element.closest("dialog[open]");
    if (closestDialog && container === document.body) container = closestDialog;
    if (!this._element.ownerDocument.documentElement.contains(this.tip)) {
      container.append(tip);
      EventHandler.trigger(this._element, this.constructor.eventName(EVENT_INSERTED));
      this._setTipListeners(tip);
    }
    await this._createFloating(tip);
    tip.classList.add(CLASS_NAME_SHOW3);
    this._setEscapeListener();
    this._setFocusOutsideListener();
    if ("ontouchstart" in document.documentElement) for (const element of document.body.children) EventHandler.on(element, "mouseover", noop);
    const complete = () => {
      EventHandler.trigger(this._element, this.constructor.eventName(EVENT_SHOWN3));
      if (this._isHovered === false) this._leave();
      this._isHovered = false;
    };
    await this._queueCallback(complete, this.tip, this._isAnimated());
  }
  async hide() {
    if (!this._isShown()) return;
    if (EventHandler.trigger(this._element, this.constructor.eventName(EVENT_HIDE3)).defaultPrevented) return;
    this._removeEscapeListener();
    this._removeFocusOutsideListener();
    this._getTipElement().classList.remove(CLASS_NAME_SHOW3);
    if ("ontouchstart" in document.documentElement) for (const element of document.body.children) EventHandler.off(element, "mouseover", noop);
    this._activeTrigger[TRIGGER_CLICK] = false;
    this._activeTrigger[TRIGGER_FOCUS] = false;
    this._activeTrigger[TRIGGER_HOVER] = false;
    this._isHovered = null;
    const complete = () => {
      if (this._isWithActiveTrigger()) return;
      if (!this._isHovered) this._disposeFloating();
      this._element.removeAttribute("aria-describedby");
      EventHandler.trigger(this._element, this.constructor.eventName(EVENT_HIDDEN5));
    };
    await this._queueCallback(complete, this.tip, this._isAnimated());
  }
  update() {
    if (this._floatingCleanup && this.tip) this._updateFloatingPosition();
  }
  _isWithContent() {
    return Boolean(this._getTitle()) || this._hasNewContent();
  }
  _hasNewContent() {
    return Boolean(this._newContent) && Object.values(this._newContent).some(Boolean);
  }
  _getTipElement() {
    if (!this.tip) this.tip = this._createTipElement(this._newContent || this._getContentForTemplate());
    return this.tip;
  }
  _createTipElement(content) {
    const tip = this._getTemplateFactory(content).toHtml();
    tip.classList.remove(CLASS_NAME_SHOW3);
    tip.classList.add(`bs-${this.constructor.NAME}-auto`);
    const tipId = getUID(this.constructor.NAME).toString();
    tip.setAttribute("id", tipId);
    if (!this._config.animation) tip.classList.add(this._getInstantClassName());
    this._setFocusTipListeners(tip);
    return tip;
  }
  setContent(content) {
    this._newContent = content;
    if (this._isShown()) {
      this._disposeFloating();
      this.show();
    }
  }
  _getTemplateFactory(content) {
    if (this._templateFactory) this._templateFactory.changeContent(content);
    else this._templateFactory = new TemplateFactory({
      ...this._config,
      content,
      extraClass: this._resolvePossibleFunction(this._config.customClass)
    });
    return this._templateFactory;
  }
  _getContentForTemplate() {
    return { [SELECTOR_TOOLTIP_INNER]: this._getTitle() };
  }
  _getTitle() {
    return this._resolvePossibleFunction(this._config.title) || this._element.getAttribute("data-bs-original-title");
  }
  _initializeOnDelegatedTarget(event) {
    return this.constructor.getOrCreateInstance(event.delegateTarget, this._getDelegateConfig());
  }
  _getInstantClassName() {
    return `${this.constructor.NAME}-instant`;
  }
  _isAnimated() {
    return getTransitionDurationFromElement(this.tip) > 0;
  }
  _isShown() {
    return this.tip && this.tip.classList.contains(CLASS_NAME_SHOW3);
  }
  _getPlacement(tip) {
    if (this._responsivePlacements) {
      const placement2 = getResponsivePlacement(this._responsivePlacements, "top");
      return AttachmentMap[placement2.toUpperCase()] || placement2;
    }
    const placement = execute(this._config.placement, [
      this,
      tip,
      this._element
    ]);
    return AttachmentMap[placement.toUpperCase()] || placement;
  }
  _parseResponsivePlacements() {
    if (typeof this._config.placement !== "string") {
      this._responsivePlacements = null;
      return;
    }
    this._responsivePlacements = parseResponsivePlacement(this._config.placement, "top");
    if (this._responsivePlacements) this._setupMediaQueryListeners();
  }
  _setupMediaQueryListeners() {
    this._disposeMediaQueryListeners();
    this._mediaQueryListeners = createBreakpointListeners(() => {
      if (this._isShown()) this._updateFloatingPosition();
    });
  }
  _disposeMediaQueryListeners() {
    disposeBreakpointListeners(this._mediaQueryListeners);
    this._mediaQueryListeners = [];
  }
  async _createFloating(tip) {
    const placement = this._getPlacement(tip);
    const arrowElement = tip.querySelector(`.${this.constructor.NAME}-arrow`);
    await this._updateFloatingPosition(tip, placement, arrowElement);
    this._floatingCleanup = autoUpdate(this._element, tip, () => this._updateFloatingPosition(tip, null, arrowElement));
  }
  async _updateFloatingPosition(tip = this.tip, placement = null, arrowElement = null) {
    if (!tip) return;
    if (!placement) placement = this._getPlacement(tip);
    if (!arrowElement) arrowElement = tip.querySelector(`.${this.constructor.NAME}-arrow`);
    const middleware = this._getFloatingMiddleware(arrowElement);
    const floatingConfig = this._getFloatingConfig(placement, middleware);
    const { x, y, placement: finalPlacement, middlewareData } = await computePosition2(this._element, tip, floatingConfig);
    Object.assign(tip.style, {
      position: "absolute",
      left: `${x}px`,
      top: `${y}px`
    });
    if (arrowElement) arrowElement.style.position = "absolute";
    Manipulator.setDataAttribute(tip, "placement", finalPlacement);
    if (arrowElement && middlewareData.arrow) {
      const { x: arrowX, y: arrowY } = middlewareData.arrow;
      const isVertical = finalPlacement.startsWith("top") || finalPlacement.startsWith("bottom");
      Object.assign(arrowElement.style, {
        left: isVertical && arrowX !== void 0 ? `${arrowX}px` : "",
        top: !isVertical && arrowY !== void 0 ? `${arrowY}px` : "",
        right: "",
        bottom: ""
      });
    }
  }
  _getOffset() {
    const { offset: offset3 } = this._config;
    if (typeof offset3 === "string") return offset3.split(",").map((value) => Number.parseInt(value, 10));
    if (typeof offset3 === "function") return ({ placement, rects }) => {
      return toFloatingOffset(offset3({
        placement,
        reference: rects.reference,
        floating: rects.floating
      }, this._element));
    };
    return offset3;
  }
  _resolvePossibleFunction(arg) {
    return execute(arg, [this._element, this._element]);
  }
  _getFloatingMiddleware(arrowElement) {
    const offsetValue = this._getOffset();
    const middleware = [
      offset2(typeof offsetValue === "function" ? offsetValue : toFloatingOffset(offsetValue)),
      flip2({ fallbackPlacements: this._config.fallbackPlacements }),
      shift2({ boundary: this._config.boundary === "clippingParents" ? "clippingAncestors" : this._config.boundary })
    ];
    if (arrowElement) middleware.push(arrow2({ element: arrowElement }));
    return middleware;
  }
  _getFloatingConfig(placement, middleware) {
    const defaultConfig = {
      placement,
      middleware
    };
    return {
      ...defaultConfig,
      ...execute(this._config.floatingConfig, [void 0, defaultConfig])
    };
  }
  _setListeners() {
    const triggers = this._config.trigger.split(" ");
    for (const trigger2 of triggers) if (trigger2 === "click") EventHandler.on(this._element, this.constructor.eventName(EVENT_CLICK), this._config.selector, (event) => {
      const context = this._initializeOnDelegatedTarget(event);
      context._activeTrigger[TRIGGER_CLICK] = !(context._isShown() && context._activeTrigger[TRIGGER_CLICK]);
      context.toggle();
    });
    else if (trigger2 !== TRIGGER_MANUAL) {
      const [eventIn, eventOut] = this._getTriggerEvents(trigger2);
      EventHandler.on(this._element, eventIn, this._config.selector, (event) => {
        const context = this._initializeOnDelegatedTarget(event);
        context._activeTrigger[event.type === "focusin" ? TRIGGER_FOCUS : TRIGGER_HOVER] = true;
        context._enter();
      });
      EventHandler.on(this._element, eventOut, this._config.selector, (event) => {
        const context = this._initializeOnDelegatedTarget(event);
        if (event.type === "focusout") context._activeTrigger[TRIGGER_FOCUS] = context._isInside(event.relatedTarget) || context._tipPointerDown;
        else context._activeTrigger[TRIGGER_HOVER] = context._isInside(event.relatedTarget);
        context._leave();
      });
    }
    this._hideModalHandler = () => {
      if (this._element) this.hide();
    };
    EventHandler.on(this._element.closest(SELECTOR_MODAL), EVENT_MODAL_HIDE, this._hideModalHandler);
  }
  _setTipListeners(tip) {
    const trigger2 = this._getTrigger();
    if (trigger2 === TRIGGER_MANUAL || trigger2.includes(TRIGGER_CLICK)) return;
    this._tipEventOut = (event) => {
      this._activeTrigger[event.type === "focusout" ? TRIGGER_FOCUS : TRIGGER_HOVER] = this._isInside(event.relatedTarget);
      this._leave();
    };
    for (const name of trigger2.split(" ")) if (name === TRIGGER_HOVER || name === TRIGGER_FOCUS) {
      const [, eventOut] = this._getTriggerEvents(name);
      EventHandler.on(tip, eventOut, this._tipEventOut);
    }
  }
  _removeTipListeners(tip) {
    if (!this._tipEventOut) return;
    const trigger2 = this._getTrigger();
    for (const name of trigger2.split(" ")) if (name === TRIGGER_HOVER || name === TRIGGER_FOCUS) {
      const [, eventOut] = this._getTriggerEvents(name);
      EventHandler.off(tip, eventOut, this._tipEventOut);
    }
    this._tipEventOut = null;
  }
  _isInside(target) {
    return target instanceof Node && (this._element.contains(target) || Boolean(this.tip?.contains(target)));
  }
  _getTrigger() {
    return this._config._trigger;
  }
  _getTriggerEvents(trigger2) {
    return {
      [TRIGGER_HOVER]: [this.constructor.eventName(EVENT_MOUSEENTER), this.constructor.eventName(EVENT_MOUSELEAVE)],
      [TRIGGER_FOCUS]: [this.constructor.eventName(EVENT_FOCUSIN), this.constructor.eventName(EVENT_FOCUSOUT)]
    }[trigger2];
  }
  _hasFocusTrigger() {
    return this._getTrigger().split(" ").includes(TRIGGER_FOCUS);
  }
  _setFocusTipListeners(tip) {
    if (!this._hasFocusTrigger()) return;
    EventHandler.on(tip, this.constructor.eventName(EVENT_POINTERDOWN2), () => {
      this._tipPointerDown = true;
      this._activeTrigger[TRIGGER_FOCUS] = true;
    });
    EventHandler.on(tip, this.constructor.eventName(EVENT_FOCUSIN), () => {
      this._activeTrigger[TRIGGER_FOCUS] = true;
    });
    EventHandler.on(tip, this.constructor.eventName(EVENT_FOCUSOUT), (event) => {
      this._activeTrigger[TRIGGER_FOCUS] = this._isInside(event.relatedTarget) || this._tipPointerDown;
      this._leave();
    });
  }
  _setFocusOutsideListener() {
    if (this._outsidePointerHandler || !this._hasFocusTrigger()) return;
    this._tipPointerUpHandler = () => {
      this._tipPointerDown = false;
    };
    this._outsidePointerHandler = (event) => {
      if (!this._isShown() || this._isInside(event.target)) return;
      this._activeTrigger[TRIGGER_FOCUS] = false;
      this.hide();
    };
    const doc = this._element.ownerDocument;
    doc.addEventListener(EVENT_POINTERUP2, this._tipPointerUpHandler, true);
    doc.addEventListener(EVENT_POINTERDOWN2, this._outsidePointerHandler, true);
  }
  _removeFocusOutsideListener() {
    const doc = this._element?.ownerDocument;
    if (this._tipPointerUpHandler && doc) doc.removeEventListener(EVENT_POINTERUP2, this._tipPointerUpHandler, true);
    if (this._outsidePointerHandler && doc) doc.removeEventListener(EVENT_POINTERDOWN2, this._outsidePointerHandler, true);
    this._tipPointerUpHandler = null;
    this._outsidePointerHandler = null;
    this._tipPointerDown = false;
  }
  _setEscapeListener() {
    if (this._keydownHandler) return;
    this._keydownHandler = (event) => {
      if (event.key !== ESCAPE_KEY3 || !this._isShown() || !this.tip.isConnected) return;
      event.preventDefault();
      event.stopPropagation();
      this.hide();
    };
    this._element.ownerDocument.addEventListener(EVENT_KEYDOWN, this._keydownHandler, true);
  }
  _removeEscapeListener() {
    if (!this._keydownHandler) return;
    this._element.ownerDocument.removeEventListener(EVENT_KEYDOWN, this._keydownHandler, true);
    this._keydownHandler = null;
  }
  _fixTitle() {
    const title = this._element.getAttribute("title");
    if (!title) return;
    if (!this._element.getAttribute("aria-label") && !this._element.textContent.trim()) this._element.setAttribute("aria-label", title);
    this._element.setAttribute("data-bs-original-title", title);
    this._element.removeAttribute("title");
  }
  _enter() {
    if (this._isShown() || this._isHovered) {
      this._isHovered = true;
      return Promise.resolve();
    }
    this._isHovered = true;
    return this._setTimeout(() => this._isHovered ? this.show() : void 0, this._config.delay.show);
  }
  _leave() {
    if (this._isWithActiveTrigger()) return Promise.resolve();
    this._isHovered = false;
    return this._setTimeout(() => this._isHovered ? void 0 : this.hide(), this._config.delay.hide);
  }
  _setTimeout(handler, timeout) {
    this._clearTimeout();
    return new Promise((resolve) => {
      this._resolveTimeout = resolve;
      this._timeout = setTimeout(() => {
        this._resolveTimeout = null;
        resolve(handler());
      }, timeout);
    });
  }
  _clearTimeout() {
    clearTimeout(this._timeout);
    if (this._resolveTimeout) {
      this._resolveTimeout();
      this._resolveTimeout = null;
    }
  }
  _isWithActiveTrigger() {
    return Object.values(this._activeTrigger).includes(true);
  }
  _getConfig(config) {
    const dataAttributes = Manipulator.getDataAttributes(this._element);
    for (const dataAttribute of Object.keys(dataAttributes)) if (DISALLOWED_ATTRIBUTES.has(dataAttribute)) delete dataAttributes[dataAttribute];
    config = {
      ...dataAttributes,
      ...typeof config === "object" && config ? config : {}
    };
    config = this._mergeConfigObj(config);
    config = this._configAfterMerge(config);
    this._typeCheckConfig(config);
    return config;
  }
  _configAfterMerge(config) {
    config.container = config.container === false ? document.body : getElement(config.container);
    config._trigger = config._trigger || config.trigger;
    if (typeof config.delay === "number") config.delay = {
      show: config.delay,
      hide: config.delay
    };
    if (typeof config.title === "number" || typeof config.title === "boolean") config.title = config.title.toString();
    if (typeof config.content === "number" || typeof config.content === "boolean") config.content = config.content.toString();
    return config;
  }
  _getDelegateConfig() {
    const config = {};
    for (const [key, value] of Object.entries(this._config)) if (this.constructor.Default[key] !== value) config[key] = value;
    config.selector = false;
    config.trigger = "manual";
    return config;
  }
  _disposeFloating() {
    if (this._floatingCleanup) {
      this._floatingCleanup();
      this._floatingCleanup = null;
    }
    if (this.tip) {
      this._removeTipListeners(this.tip);
      EventHandler.off(this.tip, this.constructor.EVENT_KEY);
      this.tip.remove();
      this.tip = null;
    }
  }
};
var initTooltip = (event) => {
  const target = event.target.closest(SELECTOR_DATA_TOGGLE5);
  if (!target) return;
  Tooltip.getOrCreateInstance(target);
};
EventHandler.on(document, EVENT_FOCUSIN, SELECTOR_DATA_TOGGLE5, initTooltip);
EventHandler.on(document, EVENT_MOUSEENTER, SELECTOR_DATA_TOGGLE5, initTooltip);

// <stdin>
var import_clipboard = __toESM(require_clipboard());
var themePreference = matchMedia("(prefers-color-scheme: dark)");
var chosenTheme = "system";
try {
  chosenTheme = localStorage.getItem("threadfin-theme") || "system";
} catch {
}
if (!["system", "light", "dark"].includes(chosenTheme)) chosenTheme = "system";
function applyTheme() {
  document.documentElement.dataset.bsTheme = chosenTheme === "system" ? themePreference.matches ? "dark" : "light" : chosenTheme;
}
applyTheme();
themePreference.addEventListener("change", applyTheme);
var themeControl = document.createElement("select");
themeControl.className = "tf-theme-control form-select";
themeControl.setAttribute("aria-label", "Color theme");
for (const [value, label] of [["system", "System theme"], ["light", "Light theme"], ["dark", "Dark theme"]]) {
  const option = new Option(label, value);
  themeControl.add(option);
}
themeControl.value = chosenTheme;
themeControl.addEventListener("change", () => {
  chosenTheme = themeControl.value;
  try {
    localStorage.setItem("threadfin-theme", chosenTheme);
  } catch {
  }
  applyTheme();
});
(document.querySelector(".tf-sidebar-header") || document.querySelector(".tf-auth-card header") || document.querySelector("#headline")).append(themeControl);
var ThreadfinConnection = class {
  constructor() {
    this.socket = null;
    this.queue = [];
    this.active = null;
    this.nextRequestId = 1;
    this.policyRejected = false;
    this.reconnectTimeoutId = null;
    this.preOpenFailures = 0;
    this.transportFailureInProgress = false;
  }
  enqueue(command, data) {
    var requestId = "request-" + this.nextRequestId;
    this.nextRequestId += 1;
    data["requestId"] = requestId;
    var request = {
      command,
      data,
      requestId,
      sent: false,
      settled: false,
      timeoutId: null
    };
    if (this.policyRejected) {
      request.settled = true;
      completeThreadfinRequestFailure(request, "{{.sources.transportError}}", "transport");
      return;
    }
    this.queue.push(request);
    this.pump();
  }
  connect() {
    if (this.policyRejected || this.socket !== null || this.reconnectTimeoutId !== null || this.active === null && this.queue.length == 0) {
      return;
    }
    var protocol = window.location.protocol == "https:" ? "wss://" : "ws://";
    var port = window.location.port ? ":" + window.location.port : "";
    var socket;
    try {
      socket = new WebSocket(protocol + window.location.hostname + port + "/data/");
    } catch (_error) {
      console.warn("WebSocket connection failed");
      this.settlePreOpenFailure(null);
      return;
    }
    this.socket = socket;
    var opened = false;
    socket.onopen = () => {
      if (this.socket !== socket || this.policyRejected) {
        return;
      }
      opened = true;
      this.preOpenFailures = 0;
      WS_AVAILABLE = true;
      console.log("WebSocket connection opened");
      this.pump();
    };
    socket.onmessage = (event) => {
      if (this.socket !== socket || this.policyRejected) {
        return;
      }
      var response;
      try {
        response = JSON.parse(event.data);
      } catch (_error) {
        this.settleProtocolFailure();
        return;
      }
      this.settleResponse(response);
    };
    socket.onerror = () => {
      if (this.socket !== socket || this.policyRejected) {
        return;
      }
      if (!opened) {
        this.settlePreOpenFailure(socket);
        return;
      }
      this.settleTransportFailure();
    };
    socket.onclose = (event) => {
      if (event.code == 1008) {
        this.rejectPolicyClose(socket);
        return;
      }
      if (this.socket !== socket || this.policyRejected) {
        return;
      }
      if (!opened) {
        this.settlePreOpenFailure(socket);
        return;
      }
      this.settleTransportFailure();
    };
  }
  pump() {
    if (this.policyRejected || this.transportFailureInProgress || this.active !== null || this.queue.length == 0 || this.reconnectTimeoutId !== null) {
      return;
    }
    if (this.socket === null) {
      this.connect();
      return;
    }
    if (this.socket.readyState !== WebSocket.OPEN) {
      return;
    }
    var request = this.queue.shift();
    this.active = request;
    request.sent = true;
    request.timeoutId = setTimeout(() => {
      if (this.active !== request || request.settled) {
        return;
      }
      this.settleTransportFailure();
    }, 3e4);
    try {
      this.socket.send(JSON.stringify(request.data));
    } catch (_error) {
      this.settleTransportFailure();
    }
  }
  settleResponse(response) {
    var responseIsObject = response && typeof response == "object" && !Array.isArray(response);
    if (this.active === null || !responseIsObject || response["requestId"] !== this.active.requestId) {
      this.settleProtocolFailure();
      return;
    }
    var request = this.active;
    if (request.settled) {
      return;
    }
    request.settled = true;
    clearTimeout(request.timeoutId);
    request.timeoutId = null;
    this.active = null;
    showElement("loading", false);
    console.log("WebSocket response received:", request.command);
    completeThreadfinRequestResponse(request, response);
    this.pump();
  }
  settleTransportFailure() {
    var socket = this.socket;
    this.socket = null;
    this.transportFailureInProgress = true;
    this.settleActiveFailure("{{.sources.transportError}}", "transport");
    this.closeSocket(socket);
    this.transportFailureInProgress = false;
    this.scheduleReconnect();
  }
  settleProtocolFailure() {
    var socket = this.socket;
    this.socket = null;
    this.transportFailureInProgress = true;
    this.settleActiveFailure("{{.sources.responseInvalid}}", "transport");
    this.closeSocket(socket);
    this.transportFailureInProgress = false;
    this.scheduleReconnect();
  }
  settlePreOpenFailure(socket) {
    if (this.policyRejected || socket !== null && this.socket !== socket) {
      return;
    }
    if (socket !== null) {
      this.socket = null;
      this.closeSocket(socket);
    }
    this.preOpenFailures += 1;
    if (this.preOpenFailures >= 2) {
      this.rejectTerminalFailure(socket, true);
      return;
    }
    this.scheduleReconnect();
  }
  settleActiveFailure(message, failureKind) {
    var request = this.active;
    if (request === null || request.settled) {
      return;
    }
    request.settled = true;
    clearTimeout(request.timeoutId);
    request.timeoutId = null;
    this.active = null;
    showElement("loading", false);
    console.warn("WebSocket request failed:", request.command);
    completeThreadfinRequestFailure(request, message, failureKind);
  }
  closeSocket(socket) {
    if (socket && socket.readyState !== WebSocket.CLOSED) {
      socket.close();
    }
  }
  scheduleReconnect() {
    if (this.policyRejected || this.reconnectTimeoutId !== null || this.active !== null || this.queue.length == 0) {
      return;
    }
    this.reconnectTimeoutId = setTimeout(() => {
      this.reconnectTimeoutId = null;
      this.pump();
    }, 250);
  }
  rejectPolicyClose(closedSocket) {
    this.rejectTerminalFailure(closedSocket, false);
  }
  rejectTerminalFailure(closedSocket, showUnavailableAlert) {
    if (this.policyRejected) {
      return;
    }
    this.policyRejected = true;
    clearTimeout(this.reconnectTimeoutId);
    this.reconnectTimeoutId = null;
    this.preOpenFailures = 0;
    var currentSocket = this.socket;
    this.socket = null;
    this.settleActiveFailure("{{.sources.transportError}}", "transport");
    var pending = this.queue;
    this.queue = [];
    pending.forEach((request) => {
      if (!request.settled) {
        request.settled = true;
        completeThreadfinRequestFailure(request, "{{.sources.transportError}}", "transport");
      }
    });
    if (currentSocket !== closedSocket) {
      this.closeSocket(currentSocket);
    }
    if (showUnavailableAlert) {
      alert("No websocket connection to Threadfin could be established. Check your network configuration.");
    }
    location.reload();
  }
};
var THREADFIN_CONNECTION = new ThreadfinConnection();
var Server = class {
  constructor(cmd) {
    this.cmd = cmd;
  }
  request(data) {
    if (this.cmd != "updateLog") {
      UNDO = new Object();
    }
    var requestData = Object.assign({}, data);
    requestData["cmd"] = this.cmd;
    THREADFIN_CONNECTION.enqueue(this.cmd, requestData);
  }
};
function completeThreadfinRequestResponse(request, response) {
  var command = request.command;
  var data = request.data;
  if (typeof completeSourceRequest == "function") {
    completeSourceRequest(command, data, response);
  }
  if (typeof completeFilterRequest == "function") {
    completeFilterRequest(command, data, response);
  }
  if (command == "saveWizard" && typeof completeConfigurationWizardRequest == "function") {
    completeConfigurationWizardRequest(response);
  }
  if (typeof completeMappingRequest == "function") {
    completeMappingRequest(command, data, response);
  }
  applyThreadfinResponse(command, data, response);
}
function completeThreadfinRequestFailure(request, message, failureKind) {
  var command = request.command;
  var data = request.data;
  var response = { status: false, err: message, requestId: request.requestId };
  if (typeof completeSourceRequest == "function") {
    completeSourceRequest(command, data, response);
  }
  if (typeof completeFilterRequest == "function") {
    completeFilterRequest(command, data, response);
  }
  if (command == "saveWizard" && typeof completeConfigurationWizardRequest == "function") {
    completeConfigurationWizardRequest(response);
  }
  if (typeof completeMappingRequest == "function") {
    completeMappingRequest(command, data, response, failureKind);
  }
}
function applyThreadfinResponse(command, data, response) {
  var responseIsObject = response && typeof response == "object" && !Array.isArray(response);
  if (!responseIsObject || response["status"] !== true) {
    if (command == "saveEpgMapping") {
      if (responseIsObject && response["xepg"] && response["xepg"]["epgMapping"] && SERVER["xepg"]) {
        SERVER["xepg"]["epgMapping"] = response["xepg"]["epgMapping"];
      }
      return;
    }
    if (responseIsObject && response["status"] === false) {
      alert(response["err"] || "{{.sources.responseInvalid}}");
      if (response.hasOwnProperty("reload")) {
        location.reload();
      }
    } else {
      alert("{{.sources.responseInvalid}}");
    }
    return;
  }
  if (response.hasOwnProperty("token")) {
    document.cookie = "Token=" + response["token"];
  }
  if (response.hasOwnProperty("probeInfo")) {
    var probeDetails = document.getElementById("probeDetails");
    if (probeDetails) {
      if (response["probeInfo"]["resolution"] !== void 0) {
        probeDetails.innerHTML = "";
        appendProbeDetail(probeDetails, "Resolution", response["probeInfo"]["resolution"], "");
        appendProbeDetail(probeDetails, "Frame Rate", response["probeInfo"]["frameRate"], " FPS");
        appendProbeDetail(probeDetails, "Audio", response["probeInfo"]["audioChannel"], "");
      }
    }
  }
  if (response.hasOwnProperty("logoURL")) {
    var div = document.getElementById("channel-icon");
    div.value = response["logoURL"];
    div.className = "changed";
    return;
  }
  switch (command) {
    case "updateLog":
      mergeUpdateLogResponse(response);
      refreshOverviewOperationalState(SERVER);
      refreshActivityOperationalState(SERVER);
      if (document.getElementById("content_log")) {
        showLogs(false);
      }
      return;
    default:
      SERVER = new Object();
      SERVER = response;
      break;
  }
  if (response.hasOwnProperty("openMenu")) {
    openLegacyMenu(response["openMenu"]);
    showElement("popup", false);
  }
  if (response.hasOwnProperty("openLink")) {
    window.location = response["openLink"];
  }
  if (response.hasOwnProperty("alert")) {
    alert(response["alert"]);
  }
  if (response.hasOwnProperty("reload")) {
    if (command == "saveWizard" && typeof completeConfigurationWizard == "function") {
      completeConfigurationWizard();
      return;
    }
    location.reload();
  }
  if (response.hasOwnProperty("wizard")) {
    createLayout();
    showConfigurationWizard(response["wizard"]);
    return;
  }
  createLayout();
}
function appendProbeDetail(container, label, value, suffix) {
  var row = document.createElement("p");
  row.appendChild(document.createTextNode(label + ": "));
  var detail = document.createElement("span");
  detail.className = "text-primary";
  detail.textContent = String(value) + suffix;
  row.appendChild(detail);
  container.appendChild(row);
}
function mergeUpdateLogResponse(response) {
  var server = overviewRecord(SERVER);
  var update = overviewRecord(response);
  mergeUpdateLogRecord(server, "clientInfo", update.clientInfo);
  mergeUpdateLogRecord(server, "log", update.log);
}
function mergeUpdateLogRecord(server, key, updateValue) {
  var existing = overviewRecord(server[key]);
  if (server[key] !== existing) {
    server[key] = existing;
  }
  var update = overviewRecord(updateValue);
  Object.keys(update).forEach((updateKey) => {
    existing[updateKey] = update[updateKey];
  });
}
var mappingDummyPrograms = [
  "PPV",
  "30_Minutes",
  "60_Minutes",
  "90_Minutes",
  "120_Minutes",
  "180_Minutes",
  "240_Minutes",
  "360_Minutes"
];
function mappingRecord(value) {
  return value && typeof value == "object" && !Array.isArray(value) ? value : {};
}
function mappingString(value) {
  return value === void 0 || value === null ? "" : String(value);
}
function mappingDeepClone(value) {
  if (value === void 0 || value === null) {
    return value;
  }
  return JSON.parse(JSON.stringify(value));
}
function mappingComparable(value) {
  if (Array.isArray(value)) {
    return "[" + value.map(mappingComparable).join(",") + "]";
  }
  if (value && typeof value == "object") {
    return "{" + Object.keys(value).sort().map(function(key) {
      return JSON.stringify(key) + ":" + mappingComparable(value[key]);
    }).join(",") + "}";
  }
  return JSON.stringify(value);
}
function mappingAuthoritativeMap(server) {
  return mappingRecord(mappingRecord(mappingRecord(server).xepg).epgMapping);
}
function mappingXMLTVMap(server) {
  return mappingRecord(mappingRecord(mappingRecord(server).xepg).xmltvMap);
}
function createMappingWorkspaceState(server) {
  var baseline = mappingDeepClone(mappingAuthoritativeMap(server));
  var xmltvMap = mappingDeepClone(mappingXMLTVMap(server));
  var draft = mappingDeepClone(baseline);
  var hasAttention = Object.keys(draft).some(function(id) {
    return mappingAttentionReasons(draft[id], xmltvMap).length > 0;
  });
  return {
    baseline,
    draft,
    xmltvMap,
    dirtyIDs: /* @__PURE__ */ new Set(),
    selected: /* @__PURE__ */ new Set(),
    segment: hasAttention ? "attention" : "active",
    selectionAnchor: "",
    saveState: "idle",
    feedback: ""
  };
}
function mappingAttentionReasons(channelValue, xmltvValue) {
  var channel = mappingRecord(channelValue);
  var xmltvMap = mappingRecord(xmltvValue);
  var file = mappingString(channel["x-xmltv-file"]);
  var program = mappingString(channel["x-mapping"]);
  var reasons = [];
  var missing = !file || file == "-" || !program || program == "-";
  if (missing) {
    reasons.push("missing");
  } else if (file == "Threadfin Dummy") {
    if (mappingDummyPrograms.indexOf(program) == -1) {
      reasons.push("invalid");
    }
  } else {
    var guide = mappingRecord(xmltvMap[file]);
    if (!Object.prototype.hasOwnProperty.call(xmltvMap, file) || !Object.prototype.hasOwnProperty.call(guide, program)) {
      reasons.push("invalid");
    }
  }
  if (channel["x-hide-channel"] === true) {
    reasons.push("hidden");
  }
  if (channel["x-active"] !== true) {
    reasons.push("inactive");
  }
  return reasons;
}
function mappingSearchText(channel, reasons) {
  var values = [];
  function collect(value) {
    if (value === void 0 || value === null) {
      return;
    }
    if (Array.isArray(value)) {
      value.forEach(collect);
    } else if (typeof value == "object") {
      Object.keys(value).forEach(function(key) {
        collect(value[key]);
      });
    } else {
      values.push(String(value));
    }
  }
  collect(channel);
  reasons.forEach(function(reason) {
    values.push(mappingAttentionReasonLabel(reason));
  });
  return values.join(" ").toLocaleLowerCase();
}
function mappingAttentionReasonLabel(reason) {
  switch (reason) {
    case "missing":
      return "Missing EPG assignment";
    case "invalid":
      return "Invalid EPG assignment";
    case "hidden":
      return "Hidden from outputs";
    case "inactive":
      return "Inactive";
  }
}
function mappingVisibleRows(state, queryValue) {
  var query = queryValue || {};
  var segment = query.segment || state.segment;
  var search = mappingString(query.search).trim().toLocaleLowerCase();
  var rows = [];
  Object.keys(state.draft).forEach(function(id, originalIndex) {
    var channel = mappingRecord(state.draft[id]);
    var reasons = mappingAttentionReasons(channel, state.xmltvMap);
    if (segment == "attention" && reasons.length == 0) {
      return;
    }
    if (segment == "active" && channel["x-active"] !== true) {
      return;
    }
    if (segment == "inactive" && channel["x-active"] === true) {
      return;
    }
    if (query.playlist && mappingString(channel["_file.m3u.id"]) != query.playlist) {
      return;
    }
    if (query.group && mappingString(channel["x-group-title"] || channel["group-title"]) != query.group) {
      return;
    }
    if (query.xmltv && mappingString(channel["x-xmltv-file"]) != query.xmltv) {
      return;
    }
    if (query.activation == "active" && channel["x-active"] !== true) {
      return;
    }
    if (query.activation == "inactive" && channel["x-active"] === true) {
      return;
    }
    if (query.reason && reasons.indexOf(query.reason) == -1) {
      return;
    }
    if (search && mappingSearchText(channel, reasons).indexOf(search) == -1) {
      return;
    }
    rows.push({ id, channel, reasons, originalIndex });
  });
  return mappingSortRows(rows, query.sort || "number", query.descending === true);
}
function mappingSortRows(rows, sort, descending) {
  var sorted = rows.slice();
  var direction = descending ? -1 : 1;
  sorted.sort(function(left, right) {
    var comparison = 0;
    if (sort == "number") {
      var leftNumber = Number.parseFloat(mappingString(left.channel["x-channelID"]));
      var rightNumber = Number.parseFloat(mappingString(right.channel["x-channelID"]));
      var leftValid = Number.isFinite(leftNumber);
      var rightValid = Number.isFinite(rightNumber);
      comparison = leftValid && rightValid ? leftNumber - rightNumber : leftValid ? -1 : rightValid ? 1 : mappingString(left.channel["x-channelID"]).localeCompare(mappingString(right.channel["x-channelID"]));
    } else {
      var key = sort == "name" ? "x-name" : sort == "playlist" ? "_file.m3u.id" : sort == "group" ? "x-group-title" : "x-xmltv-file";
      comparison = mappingString(left.channel[key]).localeCompare(mappingString(right.channel[key]), void 0, { numeric: true, sensitivity: "base" });
    }
    return comparison == 0 ? left.originalIndex - right.originalIndex : comparison * direction;
  });
  return sorted;
}
function mappingSetSelected(state, id, selected) {
  if (!Object.prototype.hasOwnProperty.call(state.draft, id)) {
    return;
  }
  if (selected) {
    state.selected.add(id);
  } else {
    state.selected.delete(id);
  }
  state.selectionAnchor = id;
}
function mappingSelectRange(state, visibleIDs, id, selected, shift3) {
  var target = visibleIDs.indexOf(id);
  var anchor = visibleIDs.indexOf(state.selectionAnchor);
  if (shift3 && target >= 0 && anchor >= 0) {
    var start = Math.min(target, anchor);
    var end = Math.max(target, anchor);
    for (var index = start; index <= end; index++) {
      if (selected)
        state.selected.add(visibleIDs[index]);
      else
        state.selected.delete(visibleIDs[index]);
    }
    state.selectionAnchor = id;
    return;
  }
  mappingSetSelected(state, id, selected);
}
function mappingSelectVisible(state, visibleIDs, selected) {
  visibleIDs.forEach(function(id) {
    if (selected)
      state.selected.add(id);
    else
      state.selected.delete(id);
  });
}
function mappingSelectedVisibleCount(state, visibleIDs) {
  return visibleIDs.reduce(function(count, id) {
    return count + (state.selected.has(id) ? 1 : 0);
  }, 0);
}
function mappingSelectedVisibleIDs(state, visibleRows) {
  return visibleRows.filter(function(row) {
    return state.selected.has(row.id);
  }).map(function(row) {
    return row.id;
  });
}
function mappingDraftIsLocked(state) {
  return state.saveState != "idle";
}
function mappingDirtyIDs(state) {
  var ids = new Set(Object.keys(state.baseline).concat(Object.keys(state.draft)));
  var dirty = [];
  ids.forEach(function(id) {
    if (mappingComparable(state.baseline[id]) != mappingComparable(state.draft[id])) {
      dirty.push(id);
    }
  });
  state.dirtyIDs = new Set(dirty);
  return dirty;
}
function mappingResolveChannelNumber(state, value) {
  var number = Number.parseFloat(mappingString(value));
  if (!Number.isFinite(number)) {
    return void 0;
  }
  var used = Object.keys(state.draft).map(function(id) {
    return Number.parseFloat(mappingString(mappingRecord(state.draft[id])["x-channelID"]));
  }).filter(Number.isFinite);
  while (used.indexOf(number) != -1) {
    number = Math.floor(number) == number ? number + 1 : Math.round((number + 0.1) * 10) / 10;
  }
  return String(number);
}
function mappingApplyChannelPatch(state, ids, patchValue, optionsValue) {
  if (mappingDraftIsLocked(state)) {
    return { ok: false, error: "Wait for the current save result before editing the draft" };
  }
  var patch = mappingRecord(patchValue);
  var options = optionsValue || {};
  var sequential;
  if (options.sequentialStart !== void 0) {
    sequential = Number.parseFloat(options.sequentialStart);
    if (!Number.isFinite(sequential)) {
      return { ok: false, error: "Invalid channel number" };
    }
  }
  var resolvedNumber;
  if (ids.length == 1 && options.numberChanged && Object.prototype.hasOwnProperty.call(patch, "x-channelID")) {
    resolvedNumber = mappingResolveChannelNumber(state, patch["x-channelID"]);
    if (resolvedNumber === void 0) {
      return { ok: false, error: "Invalid channel number" };
    }
  }
  ids.forEach(function(id, index) {
    if (!Object.prototype.hasOwnProperty.call(state.draft, id)) {
      return;
    }
    var channel = state.draft[id];
    Object.keys(patch).forEach(function(key) {
      if (key == "x-channelID" && ids.length > 1) {
        return;
      }
      channel[key] = mappingDeepClone(patch[key]);
    });
    if (resolvedNumber !== void 0) {
      channel["x-channelID"] = resolvedNumber;
    }
    if (options.sequentialStart !== void 0) {
      channel["x-channelID"] = String(sequential + index);
    }
    if (channel["x-xmltv-file"] == "-" || channel["x-mapping"] == "-") {
      channel["x-active"] = false;
    }
  });
  mappingDirtyIDs(state);
  return { ok: true };
}
function mappingAssignDummy(state, ids, program) {
  if (mappingDummyPrograms.indexOf(program) == -1) {
    return { ok: false, error: "Invalid dummy guide" };
  }
  return mappingApplyChannelPatch(state, ids, {
    "x-xmltv-file": "Threadfin Dummy",
    "x-mapping": program,
    "x-active": true,
    "x-update-channel-icon": false
  });
}
function mappingRevertDraft(state) {
  if (mappingDraftIsLocked(state)) {
    return;
  }
  state.draft = mappingDeepClone(state.baseline);
  state.dirtyIDs.clear();
  state.saveState = "idle";
  state.feedback = "";
}
function mappingReconcileAuthoritative(state, authoritativeValue, submittedValue, successful) {
  var authoritative = mappingDeepClone(mappingRecord(authoritativeValue));
  var previousBaseline = mappingDeepClone(state.baseline);
  var localDraft = mappingDeepClone(state.draft);
  var authoritativeComparable = mappingComparable(authoritative);
  var submittedComparable = submittedValue === void 0 ? "" : mappingComparable(mappingRecord(submittedValue));
  var persisted = successful === true || authoritativeComparable == submittedComparable;
  state.baseline = mappingDeepClone(authoritative);
  if (persisted) {
    state.draft = mappingDeepClone(authoritative);
  } else {
    state.draft = mappingDeepClone(authoritative);
    var ids = new Set(Object.keys(previousBaseline).concat(Object.keys(localDraft)));
    ids.forEach(function(id) {
      if (mappingComparable(previousBaseline[id]) == mappingComparable(localDraft[id])) {
        return;
      }
      if (mappingComparable(authoritative[id]) == mappingComparable(localDraft[id])) {
        return;
      }
      if (Object.prototype.hasOwnProperty.call(localDraft, id)) {
        state.draft[id] = mappingDeepClone(localDraft[id]);
      } else {
        delete state.draft[id];
      }
    });
    persisted = mappingDirtyIDs(state).length == 0 && authoritativeComparable != mappingComparable(previousBaseline);
  }
  mappingDirtyIDs(state);
  state.selected.forEach(function(id) {
    if (!Object.prototype.hasOwnProperty.call(state.draft, id))
      state.selected.delete(id);
  });
  return { persisted };
}
var filterPageFeedback = {};
var filterPopupInvoker;
var filterPopupFocusKey = "";
var filterPopupFocusListenerAttached = false;
var filterFocusAfterDeletion = false;
var filterDeletionFallback;
function filterRecord(value) {
  return value && typeof value == "object" && !Array.isArray(value) ? value : {};
}
function filterString(value) {
  return typeof value == "string" ? value : value === void 0 || value === null ? "" : String(value);
}
function filterTerms(value) {
  return filterString(value).split(",").map(function(term) {
    return term.trim();
  }).filter(function(term) {
    return term.length > 0;
  });
}
function filterQuotedTerms(value) {
  return filterTerms(value).map(function(term) {
    return '"' + term + '"';
  }).join(" or ");
}
function filterRuleSummary(filter) {
  var values = filterRecord(filter);
  var caseDescription = values.caseSensitive === true ? "case-sensitive" : "case-insensitive";
  if (values.type == "custom-filter") {
    return filterCustomRuleSummary(filterString(values.filter), caseDescription);
  }
  var summary = 'Matches streams whose group is exactly "' + filterString(values.filter) + '" (' + caseDescription + ").";
  var include = filterQuotedTerms(values.include);
  if (include) {
    summary += " Includes stream names containing any of " + include + ".";
  }
  var exclude = filterQuotedTerms(values.exclude);
  if (exclude) {
    summary += " Excludes stream names containing any of " + exclude + ".";
  }
  return summary;
}
function filterCanonicalCustomTerms(value) {
  if (!value || value.indexOf(".") != -1) {
    return void 0;
  }
  var terms = value.split(",").map(function(term) {
    return term.trim();
  });
  return terms.every(function(term) {
    return term.length > 0;
  }) ? terms : void 0;
}
function filterQuotedCanonicalTerms(terms) {
  return terms.map(function(term) {
    return '"' + term + '"';
  }).join(" or ");
}
function filterCustomRuleSummary(rule, caseDescription) {
  var parsed = /^([^\s{}!](?:[^{}!]*[^\s{}!])?)(?: \{([^{}!]+)\})?(?: !\{([^{}!]+)\})?$/.exec(rule);
  var includeTerms = parsed ? filterCanonicalCustomTerms(parsed[2]) : void 0;
  var excludeTerms = parsed ? filterCanonicalCustomTerms(parsed[3]) : void 0;
  if (!parsed || parsed[2] !== void 0 && !includeTerms || parsed[3] !== void 0 && !excludeTerms) {
    return "The custom rule will be evaluated as entered against complete stream data (" + caseDescription + "). Its include/exclude clauses are not summarized.";
  }
  var summary = 'Matches complete stream data containing "' + parsed[1] + '" (' + caseDescription + ").";
  var include = includeTerms ? filterQuotedCanonicalTerms(includeTerms) : "";
  if (include) {
    summary += " Includes complete stream data containing any of " + include + ".";
  }
  var exclude = excludeTerms ? filterQuotedCanonicalTerms(excludeTerms) : "";
  if (exclude) {
    summary += " Excludes complete stream data containing any of " + exclude + ".";
  }
  return summary;
}
function filterStreamCounts(server) {
  var data = filterRecord(filterRecord(server).data);
  var preview = filterRecord(data.StreamPreviewUI);
  var selected = Array.isArray(preview.activeStreams) ? preview.activeStreams.length : 0;
  var excluded = Array.isArray(preview.inactiveStreams) ? preview.inactiveStreams.length : 0;
  return { imported: selected + excluded, selected, excluded };
}
function selectFilterList(server) {
  var settings = filterRecord(filterRecord(server).settings);
  var saved = filterRecord(settings.filter);
  var filters = [];
  Object.keys(saved).forEach(function(id) {
    if (id == "-1") {
      return;
    }
    var filter = filterRecord(saved[id]);
    if (filter.type != "group-title" && filter.type != "custom-filter") {
      return;
    }
    filters.push({
      id,
      type: filter.type,
      name: filterString(filter.name) || "{{.filter.unnamed}}",
      startingNumber: filterString(filter.startingNumber),
      rule: filterString(filter.filter),
      summary: filterRuleSummary(filter)
    });
  });
  filters.sort(function(left, right) {
    var typeOrder = left.type == right.type ? 0 : left.type == "group-title" ? -1 : 1;
    if (typeOrder != 0) {
      return typeOrder;
    }
    var nameOrder = left.name.localeCompare(right.name);
    return nameOrder == 0 ? left.id.localeCompare(right.id) : nameOrder;
  });
  return filters;
}
function renderFilterManagementPage(host) {
  host.innerHTML = "";
  var page = document.createElement("div");
  page.className = "tf-filters";
  var header = document.createElement("header");
  header.className = "tf-filters-header";
  var titles = document.createElement("div");
  var title = document.createElement("h1");
  title.id = "filter-heading";
  title.tabIndex = -1;
  title.textContent = "Filter";
  var purpose = document.createElement("p");
  purpose.textContent = "Filters select which imported streams become channels available for Mapping.";
  titles.appendChild(title);
  titles.appendChild(purpose);
  header.appendChild(titles);
  var addGroup = createFilterAction("{{.filter.addGroup}}", "add-group", function(button) {
    openFilterPopup("group-title", void 0, button);
  });
  header.appendChild(addGroup);
  page.appendChild(header);
  var counts = filterStreamCounts(SERVER);
  var metrics = document.createElement("dl");
  metrics.className = "tf-filter-counts";
  [
    ["{{.filter.counts.imported}}", counts.imported],
    ["{{.filter.counts.selected}}", counts.selected],
    ["{{.filter.counts.excluded}}", counts.excluded]
  ].forEach(function(count) {
    var group = document.createElement("div");
    var term = document.createElement("dt");
    term.textContent = count[0];
    var value = document.createElement("dd");
    value.textContent = String(count[1]);
    group.appendChild(term);
    group.appendChild(value);
    metrics.appendChild(group);
  });
  page.appendChild(metrics);
  var feedback = document.createElement("p");
  feedback.className = "tf-filter-page-status";
  feedback.setAttribute("role", "status");
  feedback.setAttribute("aria-live", "polite");
  var pageFeedback = filterPageFeedback.filter;
  if (pageFeedback) {
    feedback.textContent = pageFeedback.message;
    feedback.setAttribute("data-state", pageFeedback.state);
  } else {
    feedback.hidden = true;
  }
  page.appendChild(feedback);
  var filters = selectFilterList(SERVER);
  if (filters.length == 0) {
    var empty = document.createElement("section");
    empty.className = "tf-filter-empty";
    var emptyTitle = document.createElement("h2");
    emptyTitle.textContent = "{{.filter.empty.title}}";
    var emptyCopy = document.createElement("p");
    emptyCopy.textContent = "{{.filter.empty.description}}";
    empty.appendChild(emptyTitle);
    empty.appendChild(emptyCopy);
    empty.appendChild(createFilterAction("{{.filter.addGroup}}", "empty-add-group", function(button) {
      openFilterPopup("group-title", void 0, button);
    }));
    page.appendChild(empty);
  } else {
    var list = document.createElement("div");
    list.className = "tf-filter-list";
    list.setAttribute("role", "list");
    filters.forEach(function(filter) {
      list.appendChild(renderFilterRow(filter));
    });
    page.appendChild(list);
  }
  var advanced = document.createElement("section");
  advanced.className = "tf-filter-advanced";
  var advancedTitle = document.createElement("h2");
  advancedTitle.textContent = "{{.filter.advanced.title}}";
  var advancedCopy = document.createElement("p");
  advancedCopy.textContent = "{{.filter.advanced.description}}";
  advanced.appendChild(advancedTitle);
  advanced.appendChild(advancedCopy);
  advanced.appendChild(createFilterAction("{{.filter.advanced.chooseType}}", "choose-type", function(button) {
    openFilterPopup("filter", void 0, button);
  }));
  advanced.appendChild(createFilterAction("{{.filter.advanced.addCustom}}", "add-custom", function(button) {
    openFilterPopup("custom-filter", void 0, button);
  }));
  page.appendChild(advanced);
  host.appendChild(page);
  if (filterFocusAfterDeletion) {
    filterDeletionFallback = addGroup || title;
  }
}
function createFilterAction(label, focusKey, listener) {
  var button = document.createElement("button");
  button.type = "button";
  button.className = "tf-filter-action";
  button.textContent = label;
  button.setAttribute("data-filter-focus-key", focusKey);
  button.addEventListener("click", function() {
    listener(button);
  });
  return button;
}
function renderFilterRow(filter) {
  var row = document.createElement("article");
  row.className = "tf-filter-row";
  row.setAttribute("role", "listitem");
  row.setAttribute("data-filter-id", filter.id);
  row.setAttribute("data-filter-type", filter.type);
  var identity = document.createElement("div");
  var name = document.createElement("h2");
  name.textContent = filter.name;
  var meta = document.createElement("p");
  meta.textContent = (filter.type == "group-title" ? "{{.filter.group}}" : "{{.filter.custom}}") + " \xB7 {{.filter.startingnumber.title}}: " + filter.startingNumber;
  var rule = document.createElement("code");
  rule.textContent = filter.rule;
  var summary = document.createElement("p");
  summary.className = "tf-filter-summary";
  summary.textContent = filter.summary;
  identity.appendChild(name);
  identity.appendChild(meta);
  identity.appendChild(rule);
  identity.appendChild(summary);
  row.appendChild(identity);
  var actions = document.createElement("div");
  actions.className = "tf-filter-actions";
  actions.appendChild(createFilterAction("{{.sources.actions.edit}}", filter.id + ":edit", function(button) {
    openFilterPopup(filter.type, { id: filter.id }, button);
  }));
  var remove = createFilterAction("{{.sources.actions.delete}}", filter.id + ":delete", function(button) {
    openFilterPopup(filter.type, { id: filter.id }, button);
    savePopupData("filter", filter.id, true, 0);
  });
  remove.className += " tf-filter-delete-action";
  actions.appendChild(remove);
  row.appendChild(actions);
  return row;
}
function openFilterPopup(dataType, element, invoker) {
  filterPopupInvoker = invoker;
  filterPopupFocusKey = invoker ? filterString(invoker.getAttribute("data-filter-focus-key")) : "";
  var modal = document.getElementById("popup");
  if (modal && !filterPopupFocusListenerAttached) {
    filterPopupFocusListenerAttached = true;
    modal.addEventListener("hidden.bs.dialog", function() {
      var target;
      if (filterFocusAfterDeletion) {
        target = filterDeletionFallback || filterPopupReplacement("add-group") || document.getElementById("filter-heading");
        filterFocusAfterDeletion = false;
        filterDeletionFallback = void 0;
      } else {
        target = filterPopupInvoker && document.contains(filterPopupInvoker) ? filterPopupInvoker : filterPopupReplacement(filterPopupFocusKey);
      }
      if (target) {
        target.focus();
      }
      filterPopupInvoker = void 0;
      filterPopupFocusKey = "";
    });
  }
  openPopUp(dataType, element);
}
function filterPopupReplacement(focusKey) {
  if (!focusKey) {
    return void 0;
  }
  var candidates = document.querySelectorAll("[data-filter-focus-key]");
  for (var index = 0; index < candidates.length; index++) {
    var candidate = candidates[index];
    if (candidate.getAttribute("data-filter-focus-key") == focusKey) {
      return candidate;
    }
  }
  return void 0;
}
function enhanceFilterPopup(dataType) {
  var popup = document.getElementById("popup-custom");
  if (!popup) {
    return;
  }
  popup.classList.remove("tf-filter-popup");
  if (dataType != "group-title" && dataType != "custom-filter") {
    return;
  }
  popup.classList.add("tf-filter-popup");
  var fields = popup.querySelectorAll("input, select");
  Array.prototype.forEach.call(fields, function(field) {
    var row = field.closest("tr");
    var title = row ? row.querySelector("td:first-child") : void 0;
    if (title && title.textContent) {
      field.setAttribute("aria-label", title.textContent.replace(/:\s*$/, ""));
    }
  });
  ["name", "filter"].forEach(function(fieldName) {
    var field = popup.querySelector('[name="' + fieldName + '"]');
    if (!field || !field.parentElement) {
      return;
    }
    field.setAttribute("aria-required", "true");
    var error = document.createElement("p");
    error.id = "filter-" + fieldName + "-error";
    error.className = "tf-filter-field-error";
    error.hidden = true;
    field.setAttribute("aria-errormessage", error.id);
    field.parentElement.appendChild(error);
  });
  var example = document.createElement("p");
  example.className = "tf-filter-field-help";
  example.textContent = "{{.filter.example}}";
  popup.appendChild(example);
  var summary = document.createElement("p");
  summary.id = "filter-rule-summary";
  summary.className = "tf-filter-rule-summary";
  summary.setAttribute("role", "status");
  summary.setAttribute("aria-live", "polite");
  popup.appendChild(summary);
  var status = document.createElement("p");
  status.id = "filter-form-status";
  status.className = "tf-filter-form-status";
  status.setAttribute("role", "status");
  status.setAttribute("aria-live", "polite");
  status.hidden = true;
  popup.appendChild(status);
  var update = function() {
    filterUpdatePopupSummary();
  };
  Array.prototype.forEach.call(fields, function(field) {
    field.addEventListener("input", update);
    field.addEventListener("change", update);
  });
  filterUpdatePopupSummary();
}
function filterPopupValues() {
  var popup = document.getElementById("popup-custom");
  var values = {};
  if (!popup) {
    return values;
  }
  var fields = popup.querySelectorAll("input, select");
  Array.prototype.forEach.call(fields, function(field) {
    if (!field.name) {
      return;
    }
    values[field.name] = field.type == "checkbox" ? field.checked : field.value;
  });
  return values;
}
function filterUpdatePopupSummary() {
  var summary = document.getElementById("filter-rule-summary");
  if (summary) {
    summary.textContent = filterRuleSummary(filterPopupValues());
  }
}
function validateFilterPopup(dataType) {
  if (dataType != "group-title" && dataType != "custom-filter") {
    return true;
  }
  var popup = document.getElementById("popup-custom");
  if (!popup) {
    return false;
  }
  filterClearPopupErrors();
  var name = popup.querySelector('[name="name"]');
  if (!name || !name.value.trim()) {
    filterSetPopupError(name, "filter-name-error", "{{.filter.validation.nameRequired}}");
    return false;
  }
  var rule = popup.querySelector('[name="filter"]');
  if (!rule || !rule.value.trim()) {
    filterSetPopupError(rule, "filter-filter-error", "{{.filter.validation.ruleRequired}}");
    return false;
  }
  return true;
}
function filterClearPopupErrors() {
  var popup = document.getElementById("popup-custom");
  if (!popup) {
    return;
  }
  Array.prototype.forEach.call(popup.querySelectorAll('[aria-invalid="true"]'), function(field) {
    field.removeAttribute("aria-invalid");
  });
  Array.prototype.forEach.call(popup.querySelectorAll(".tf-filter-field-error"), function(error) {
    error.textContent = "";
    error.hidden = true;
  });
}
function filterSetPopupError(field, errorID, message) {
  if (field) {
    field.setAttribute("aria-invalid", "true");
  }
  var error = document.getElementById(errorID);
  if (error) {
    error.textContent = message;
    error.hidden = false;
  }
  if (field) {
    field.focus();
  }
}
function filterSetFormStatus(message, state) {
  var status = document.getElementById("filter-form-status");
  if (!status) {
    return;
  }
  status.textContent = message;
  status.setAttribute("data-state", state);
  status.hidden = false;
}
function beginFilterRequest(dataType, id, remove) {
  if (dataType != "filter") {
    return;
  }
  filterSetFormStatus(remove == true ? "{{.filter.feedback.deleting}}" : "{{.filter.feedback.saving}}", "progress");
}
function completeFilterRequest(command, data, response) {
  if (command != "saveFilter") {
    return;
  }
  var result = filterRecord(response);
  if (result.status !== true) {
    var error = result.status === false ? filterString(result.err) || "{{.filter.feedback.error}}" : "{{.sources.responseInvalid}}";
    filterSetFormStatus(error, "error");
    filterPageFeedback.filter = { state: "error", message: error };
    return;
  }
  var request = filterRecord(filterRecord(data).filter);
  var ids = Object.keys(request);
  var deleted = ids.length == 1 && filterRecord(request[ids[0]]).delete === true;
  if (deleted) {
    filterFocusAfterDeletion = true;
  }
  filterPageFeedback.filter = { state: "success", message: deleted ? "{{.filter.feedback.deleted}}" : "{{.filter.feedback.saved}}" };
}
var MainMenuItem = class {
  constructor(menuKey, value) {
    this.menuKey = menuKey;
    this.value = value;
  }
};
var ShowContent = class {
  constructor(menuID) {
    this.menuID = menuID;
  }
  show() {
    COLUMN_TO_SORT = -1;
    var doc = document.getElementById("content");
    if (!doc) {
      return;
    }
    doc.innerHTML = "";
    showPreview(false);
    var menuKey = menuItems[this.menuID].menuKey;
    switch (menuKey) {
      case "playlist":
      case "xmltv":
        renderSourceManagementPage(menuKey, doc);
        break;
      case "filter":
        renderFilterManagementPage(doc);
        break;
      case "mapping":
        renderMappingPage(doc);
        break;
      case "settings":
        renderSettingsPage(doc);
        break;
      case "users":
        renderUsersPage(doc);
        break;
      case "log":
        renderLogPage(doc);
        break;
      case "logout":
        var form = document.createElement("form");
        form.method = "post";
        form.action = "/web/logout";
        document.body.appendChild(form);
        form.submit();
        return;
      default:
        console.warn("Ignoring unknown legacy menu destination", menuKey);
        return;
    }
    showElement("loading", false);
  }
};
function PageReady() {
  var server = new Server("getServerConfig");
  server.request(new Object());
  setInterval(function() {
    updateLog();
  }, 1e4);
  return;
}
function createLayout() {
  var obj = SERVER["clientInfo"];
  var keys = getObjKeys(obj);
  console.log("KEYS: ", keys);
  for (var i = 0; i < keys.length; i++) {
    if (document.getElementById(keys[i])) {
      document.getElementById(keys[i]).value = obj[keys[i]];
    }
  }
  if (document.getElementById("playlist-connection-information")) {
    let activeClass = "text-primary";
    if (SERVER["clientInfo"]["activePlaylist"] / SERVER["clientInfo"]["totalPlaylist"] >= 0.6 && SERVER["clientInfo"]["activePlaylist"] / SERVER["clientInfo"]["totalPlaylist"] < 0.8) {
      activeClass = "text-warning";
    } else if (SERVER["clientInfo"]["activePlaylist"] / SERVER["clientInfo"]["totalPlaylist"] >= 0.8) {
      activeClass = "text-danger";
    }
    renderConnectionCapacity("playlist-connection-information", "Playlist Connections", SERVER["clientInfo"]["activePlaylist"], SERVER["clientInfo"]["totalPlaylist"], activeClass);
  }
  if (document.getElementById("client-connection-information")) {
    let activeClass = "text-primary";
    if (SERVER["clientInfo"]["activeClients"] / SERVER["clientInfo"]["totalClients"] >= 0.6 && SERVER["clientInfo"]["activeClients"] / SERVER["clientInfo"]["totalClients"] < 0.8) {
      activeClass = "text-warning";
    } else if (SERVER["clientInfo"]["activeClients"] / SERVER["clientInfo"]["totalClients"] >= 0.8) {
      activeClass = "text-danger";
    }
    renderConnectionCapacity("client-connection-information", "Client Connections", SERVER["clientInfo"]["activeClients"], SERVER["clientInfo"]["totalClients"], activeClass);
  }
  if (!document.getElementById("main-menu")) {
    return;
  }
  renderNavigation();
  restoreInitialDestinationFromHistory();
  return;
}
function renderConnectionCapacity(id, label, active, total, activeClass) {
  var container = document.getElementById(id);
  if (!container) {
    return;
  }
  container.innerHTML = "";
  container.appendChild(document.createTextNode(label + ": "));
  var capacity = document.createElement("span");
  capacity.className = activeClass;
  capacity.textContent = String(active) + " / " + String(total);
  container.appendChild(capacity);
}
var PopupWindow = class {
  constructor() {
    this.DocumentID = "popup-custom";
    this.InteractionID = "interaction";
    this.doc = document.getElementById(this.DocumentID);
  }
  createTitle(title) {
    var td = document.createElement("TD");
    td.className = "left";
    td.textContent = title + ":";
    return td;
  }
  createContent(element) {
    var td = document.createElement("TD");
    td.appendChild(element);
    return td;
  }
  createInteraction() {
    var div = document.createElement("div");
    div.setAttribute("id", "popup-interaction");
    div.className = "interaction";
    this.doc.appendChild(div);
  }
};
var PopupContent = class extends PopupWindow {
  constructor() {
    super(...arguments);
    this.table = document.createElement("TABLE");
    this.rowIndex = 0;
  }
  createHeadline(headline) {
    this.doc.innerHTML = "";
    var element = document.createElement("H3");
    element.textContent = headline.toUpperCase();
    this.doc.appendChild(element);
    this.table = document.createElement("TABLE");
    this.doc.appendChild(this.table);
  }
  appendRow(title, element) {
    var tr = document.createElement("TR");
    var titleCell = null;
    var rowIndex = this.rowIndex++;
    if (title.length != 0) {
      titleCell = this.createTitle(title);
      titleCell.id = "popup-field-label-" + rowIndex;
      tr.appendChild(titleCell);
    }
    tr.appendChild(this.createContent(element));
    if (titleCell) {
      var controls = [];
      var tagName = element && element.tagName ? String(element.tagName).toUpperCase() : "";
      if (tagName == "INPUT" || tagName == "SELECT" || tagName == "TEXTAREA") {
        controls.push(element);
      }
      if (element && typeof element.querySelectorAll == "function") {
        var descendants = element.querySelectorAll("input, select, textarea");
        for (var controlIndex = 0; controlIndex < descendants.length; controlIndex++) {
          if (controls.indexOf(descendants[controlIndex]) == -1) {
            controls.push(descendants[controlIndex]);
          }
        }
      }
      controls.forEach((control, controlIndex2) => {
        if (String(control.type || "").toLowerCase() == "hidden") {
          return;
        }
        if (!control.id) {
          control.id = "popup-field-" + rowIndex + "-" + controlIndex2;
        }
        if (!control.getAttribute("aria-label") && !control.getAttribute("aria-labelledby")) {
          control.setAttribute("aria-labelledby", titleCell.id);
        }
      });
    }
    this.table.appendChild(tr);
  }
  createInput(type, name, value) {
    var input = document.createElement("INPUT");
    if (value == void 0) {
      value = "";
    }
    input.setAttribute("type", type);
    input.setAttribute("name", name);
    input.setAttribute("value", value);
    return input;
  }
  createCheckbox(name) {
    var input = document.createElement("INPUT");
    input.setAttribute("type", "checkbox");
    input.setAttribute("name", name);
    return input;
  }
  createSelect(text, values, set, dbKey) {
    var select = document.createElement("SELECT");
    select.setAttribute("name", dbKey);
    for (let i = 0; i < text.length; i++) {
      var option = document.createElement("OPTION");
      option.setAttribute("value", values[i]);
      option.innerText = text[i];
      select.appendChild(option);
    }
    if (set != "") {
      select.value = set;
    }
    if (set == void 0) {
      select.value = values[0];
    }
    return select;
  }
  selectOption(select, value) {
    var s = select;
    s.options[s.selectedIndex].value = value;
    return select;
  }
  description(value, repositoryTemplate = false) {
    var span = document.createElement("PRE");
    if (repositoryTemplate) {
      appendDescriptionText(span, value);
    } else {
      span.textContent = value;
    }
    this.descriptionElement(span);
  }
  descriptionElement(element) {
    var tr = document.createElement("TR");
    var td = document.createElement("TD");
    tr.appendChild(td);
    tr.appendChild(this.createContent(element));
    this.table.appendChild(tr);
  }
  // Interaktion
  addInteraction(element) {
    var interaction = document.getElementById("popup-interaction");
    interaction.appendChild(element);
  }
};
function appendDescriptionText(element, value) {
  var lines = String(value).split(/<br\s*\/?\s*>/i);
  lines.forEach((line, index) => {
    if (index > 0) {
      element.appendChild(document.createElement("BR"));
    }
    element.appendChild(document.createTextNode(line));
  });
}
function addRuntimeEventListener(element, eventName, listener) {
  if (element && typeof element.addEventListener == "function") {
    element.addEventListener(eventName, listener);
  }
}
function openPopUp(dataType, element) {
  var data = new Object();
  var id;
  switch (element) {
    case void 0:
      switch (dataType) {
        case "group-title":
          if (id == void 0) {
            id = -1;
          }
          data = getLocalData("filter", id);
          data["type"] = "group-title";
          break;
        case "custom-filter":
          if (id == void 0) {
            id = -1;
          }
          data = getLocalData("filter", id);
          data["type"] = "custom-filter";
          break;
        default:
          data["id.provider"] = "-";
          data["type"] = dataType;
          id = "-";
          break;
      }
      break;
    default:
      id = element.id;
      data = getLocalData(dataType, id);
      break;
  }
  var content = new PopupContent();
  switch (dataType) {
    case "playlist":
      content.createHeadline("{{.playlist.playlistType.title}}");
      var text = ["M3U", "HDHomeRun"];
      var values = ["m3u", "hdhr"];
      var select = content.createSelect(text, values, "", "type");
      select.setAttribute("id", "type");
      content.appendRow("{{.playlist.type.title}}", select);
      content.createInteraction();
      var input = content.createInput("button", "cancel", "{{.button.cancel}}");
      addRuntimeEventListener(input, "click", () => showElement("popup", false));
      content.addInteraction(input);
      var input = content.createInput("button", "next", "{{.button.next}}");
      addRuntimeEventListener(input, "click", () => openPopUp(document.querySelector("#popup select[name=type]").value, void 0));
      input.setAttribute("id", "next");
      content.addInteraction(input);
      break;
    case "m3u":
      content.createHeadline(dataType);
      var dbKey = "name";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.name.placeholder}}");
      content.appendRow("{{.playlist.name.title}}", input);
      var dbKey = "description";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.description.placeholder}}");
      content.appendRow("{{.playlist.description.title}}", input);
      var dbKey = "file.source";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.fileM3U.placeholder}}");
      content.appendRow("{{.playlist.fileM3U.title}}", input);
      var text = ["-", "FFmpeg", "VLC"];
      var values = ["-", "ffmpeg", "vlc"];
      var selected = SERVER["settings"]["buffer"];
      if (data["buffer"] != void 0) {
        selected = data["buffer"];
      }
      var select = content.createSelect(text, values, selected, "buffer");
      select.setAttribute("id", "buffer");
      content.appendRow("{{.playlist.buffer.title}}", select);
      var text = new Array();
      var values = new Array();
      for (var i = 1; i <= 100; i++) {
        text.push(i.toString());
        values.push(i.toString());
      }
      var dbKey = "tuner";
      var select = content.createSelect(text, values, data[dbKey], dbKey);
      content.appendRow("{{.playlist.tuner.title}}", select);
      content.description("{{.playlist.tuner.description}}", true);
      var dbKey = "http_proxy.ip";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.http_proxy_ip.placeholder}}");
      content.appendRow("{{.playlist.http_proxy_ip.title}}", input);
      content.description("{{.playlist.http_proxy_ip.description}}", true);
      var dbKey = "http_proxy.port";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.http_proxy_port.placeholder}}");
      content.appendRow("{{.playlist.http_proxy_port.title}}", input);
      content.description("{{.playlist.http_proxy_port.description}}", true);
      var dbKey = "http_headers.origin";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.http_user_origin.placeholder}}");
      content.appendRow("{{.playlist.http_user_origin.title}}", input);
      content.description("{{.playlist.http_user_origin.description}}", true);
      var dbKey = "http_headers.referer";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.http_user_referer.placeholder}}");
      content.appendRow("{{.playlist.http_user_referer.title}}", input);
      content.description("{{.playlist.http_user_referer.description}}", true);
      content.createInteraction();
      if (data["id.provider"] != "-") {
        var input = content.createInput("button", "delete", "{{.button.delete}}");
        input.className = "delete";
        addRuntimeEventListener(input, "click", () => savePopupData("m3u", id, true, 0));
        content.addInteraction(input);
      } else {
        var input = content.createInput("button", "back", "{{.button.back}}");
        addRuntimeEventListener(input, "click", () => openPopUp("playlist", void 0));
        content.addInteraction(input);
      }
      var input = content.createInput("button", "cancel", "{{.button.cancel}}");
      addRuntimeEventListener(input, "click", () => showElement("popup", false));
      content.addInteraction(input);
      if (data["id.provider"] != "-") {
        var input = content.createInput("button", "update", "{{.button.update}}");
        addRuntimeEventListener(input, "click", () => savePopupData("m3u", id, false, 1));
        content.addInteraction(input);
      }
      var input = content.createInput("button", "save", "{{.button.save}}");
      addRuntimeEventListener(input, "click", () => savePopupData("m3u", id, false, 0));
      content.addInteraction(input);
      break;
    case "hdhr":
      content.createHeadline(dataType);
      var dbKey = "name";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.name.placeholder}}");
      content.appendRow("{{.playlist.name.title}}", input);
      var dbKey = "description";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.description.placeholder}}");
      content.appendRow("{{.playlist.description.placeholder}}", input);
      var dbKey = "file.source";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.fileHDHR.placeholder}}");
      content.appendRow("{{.playlist.fileHDHR.title}}", input);
      var text = ["-", "FFmpeg", "VLC"];
      var values = ["-", "ffmpeg", "vlc"];
      var selected = SERVER["settings"]["buffer"];
      if (data["buffer"] != void 0) {
        selected = data["buffer"];
      }
      var select = content.createSelect(text, values, selected, "buffer");
      select.setAttribute("id", "buffer");
      content.appendRow("{{.playlist.buffer.title}}", select);
      var text = new Array();
      var values = new Array();
      for (var i = 1; i <= 100; i++) {
        text.push(i.toString());
        values.push(i.toString());
      }
      var dbKey = "tuner";
      var select = content.createSelect(text, values, data[dbKey], dbKey);
      content.appendRow("{{.playlist.tuner.title}}", select);
      content.description("{{.playlist.tuner.description}}", true);
      var dbKey = "http_proxy.ip";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.http_proxy_ip.placeholder}}");
      content.appendRow("{{.playlist.http_proxy_ip.title}}", input);
      content.description("{{.playlist.http_proxy_ip.description}}", true);
      var dbKey = "http_proxy.port";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.playlist.http_proxy_port.placeholder}}");
      content.appendRow("{{.playlist.http_proxy_port.title}}", input);
      content.description("{{.playlist.http_proxy_port.description}}", true);
      content.createInteraction();
      if (data["id.provider"] != "-") {
        var input = content.createInput("button", "delete", "{{.button.delete}}");
        addRuntimeEventListener(input, "click", () => savePopupData("hdhr", id, true, 0));
        input.className = "delete";
        content.addInteraction(input);
      } else {
        var input = content.createInput("button", "back", "{{.button.back}}");
        addRuntimeEventListener(input, "click", () => openPopUp("playlist", void 0));
        content.addInteraction(input);
      }
      var input = content.createInput("button", "cancel", "{{.button.cancel}}");
      addRuntimeEventListener(input, "click", () => showElement("popup", false));
      content.addInteraction(input);
      if (data["id.provider"] != "-") {
        var input = content.createInput("button", "update", "{{.button.update}}");
        addRuntimeEventListener(input, "click", () => savePopupData("hdhr", id, false, 1));
        content.addInteraction(input);
      }
      var input = content.createInput("button", "save", "{{.button.save}}");
      addRuntimeEventListener(input, "click", () => savePopupData("hdhr", id, false, 0));
      content.addInteraction(input);
      break;
    case "filter":
      content.createHeadline(dataType);
      var dbKey = "type";
      var text = ["M3U: {{.filter.type.groupTitle}}", "Threadfin: {{.filter.type.customFilter}}"];
      var values = ["group-title", "custom-filter"];
      var select = content.createSelect(text, values, "group-title", dbKey);
      select.setAttribute("id", id);
      content.appendRow("{{.filter.type.title}}", select);
      content.createInteraction();
      var input = content.createInput("button", "cancel", "{{.button.cancel}}");
      addRuntimeEventListener(input, "click", () => showElement("popup", false));
      content.addInteraction(input);
      var input = content.createInput("button", "next", "{{.button.next}}");
      addRuntimeEventListener(input, "click", () => openPopUp(document.querySelector("#popup select[name=type]").value, void 0));
      input.setAttribute("id", "next");
      content.addInteraction(input);
      break;
    case "custom-filter":
    case "group-title":
      switch (dataType) {
        case "custom-filter":
          content.createHeadline("{{.filter.custom}}");
          break;
        case "group-title":
          content.createHeadline("{{.filter.group}}");
          break;
      }
      var dbKey = "name";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.filter.name.placeholder}}");
      content.appendRow("{{.filter.name.title}}", input);
      var dbKey = "description";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.filter.description.placeholder}}");
      content.appendRow("{{.filter.description.title}}", input);
      var dbKey = "type";
      var input = content.createInput("hidden", dbKey, data[dbKey]);
      content.appendRow("", input);
      var filterType = data[dbKey];
      switch (filterType) {
        case "custom-filter":
          var dbKey = "caseSensitive";
          var input = content.createCheckbox(dbKey);
          input.checked = data[dbKey];
          content.appendRow("{{.filter.caseSensitive.title}}", input);
          var dbKey = "filter";
          var input = content.createInput("text", dbKey, data[dbKey]);
          input.setAttribute("placeholder", "{{.filter.filterRule.placeholder}}");
          content.appendRow("{{.filter.filterRule.title}}", input);
          break;
        case "group-title":
          var dbKey = "filter";
          var groupsM3U = getLocalData("m3uGroups", "");
          var text = groupsM3U["text"];
          var values = groupsM3U["value"];
          var select = content.createSelect(text, values, data[dbKey], dbKey);
          addRuntimeEventListener(select, "change", (event) => {
            event.currentTarget.className = "changed";
          });
          content.appendRow("{{.filter.filterGroup.title}}", select);
          content.description("{{.filter.filterGroup.description}}", true);
          var dbKey = "liveEvent";
          var input = content.createCheckbox(dbKey);
          input.checked = data[dbKey];
          content.appendRow("{{.filter.liveEvent.title}}", input);
          var dbKey = "caseSensitive";
          var input = content.createCheckbox(dbKey);
          input.checked = data[dbKey];
          content.appendRow("{{.filter.caseSensitive.title}}", input);
          var dbKey = "include";
          var input = content.createInput("text", dbKey, data[dbKey]);
          input.setAttribute("placeholder", "{{.filter.include.placeholder}}");
          content.appendRow("{{.filter.include.title}}", input);
          content.description("{{.filter.include.description}}", true);
          var dbKey = "exclude";
          var input = content.createInput("text", dbKey, data[dbKey]);
          input.setAttribute("placeholder", "{{.filter.exclude.placeholder}}");
          content.appendRow("{{.filter.exclude.title}}", input);
          content.description("{{.filter.exclude.description}}", true);
          break;
        default:
          break;
      }
      var dbKey = "startingNumber";
      if (data[dbKey] !== void 0) {
        var input = content.createInput("text", dbKey, data[dbKey]);
      } else {
        var input = content.createInput("text", dbKey, "1000");
      }
      input.setAttribute("placeholder", "{{.filter.startingnumber.placeholder}}");
      content.appendRow("{{.filter.startingnumber.title}}", input);
      content.description("{{.filter.startingnumber.description}}", true);
      var dbKey = "x-category";
      var text = ["-"];
      var values = [""];
      var epgCategories = SERVER["settings"]["epgCategories"];
      var categories = epgCategories.split("|");
      for (i = 0; i <= categories.length; i++) {
        var cat = categories[i];
        if (cat) {
          var cat_split = cat.split(":");
          text.push(cat_split[0]);
          values.push(cat_split[1]);
        }
      }
      var select = content.createSelect(text, values, data[dbKey], dbKey);
      addRuntimeEventListener(select, "change", (event) => {
        event.currentTarget.className = "changed";
      });
      content.appendRow("{{.filter.category.title}}", select);
      content.createInteraction();
      var input = content.createInput("button", "delete", "{{.button.delete}}");
      addRuntimeEventListener(input, "click", () => savePopupData("filter", id, true, 0));
      input.className = "delete";
      content.addInteraction(input);
      var input = content.createInput("button", "cancel", "{{.button.cancel}}");
      addRuntimeEventListener(input, "click", () => showElement("popup", false));
      content.addInteraction(input);
      var input = content.createInput("button", "save", "{{.button.save}}");
      addRuntimeEventListener(input, "click", () => savePopupData("filter", id, false, 0));
      content.addInteraction(input);
      break;
    case "xmltv":
      content.createHeadline(dataType);
      var dbKey = "name";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.xmltv.name.placeholder}}");
      content.appendRow("{{.xmltv.name.title}}", input);
      var dbKey = "description";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.xmltv.description.placeholder}}");
      content.appendRow("{{.xmltv.description.title}}", input);
      var dbKey = "file.source";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.xmltv.fileXMLTV.placeholder}}");
      content.appendRow("{{.xmltv.fileXMLTV.title}}", input);
      var dbKey = "http_proxy.ip";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.xmltv.http_proxy_ip.placeholder}}");
      content.appendRow("{{.xmltv.http_proxy_ip.title}}", input);
      content.description("{{.xmltv.http_proxy_ip.description}}", true);
      var dbKey = "http_proxy.port";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.xmltv.http_proxy_port.placeholder}}");
      content.appendRow("{{.xmltv.http_proxy_port.title}}", input);
      content.description("{{.xmltv.http_proxy_port.description}}", true);
      content.createInteraction();
      if (data["id.provider"] != "-") {
        var input = content.createInput("button", "delete", "{{.button.delete}}");
        addRuntimeEventListener(input, "click", () => savePopupData("xmltv", id, true, 0));
        input.className = "delete";
        content.addInteraction(input);
      }
      var input = content.createInput("button", "cancel", "{{.button.cancel}}");
      addRuntimeEventListener(input, "click", () => showElement("popup", false));
      content.addInteraction(input);
      if (data["id.provider"] != "-") {
        var input = content.createInput("button", "update", "{{.button.update}}");
        addRuntimeEventListener(input, "click", () => savePopupData("xmltv", id, false, 1));
        content.addInteraction(input);
      }
      var input = content.createInput("button", "save", "{{.button.save}}");
      addRuntimeEventListener(input, "click", () => savePopupData("xmltv", id, false, 0));
      content.addInteraction(input);
      break;
    case "users":
      content.createHeadline("{{.mainMenu.item.users}}");
      var dbKey = "username";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.users.username.placeholder}}");
      content.appendRow("{{.users.username.title}}", input);
      var dbKey = "password";
      var input = content.createInput("password", dbKey, "");
      input.setAttribute("placeholder", "{{.users.password.placeholder}}");
      content.appendRow("{{.users.password.title}}", input);
      var dbKey = "confirm";
      var input = content.createInput("password", dbKey, "");
      input.setAttribute("placeholder", "{{.users.confirm.placeholder}}");
      content.appendRow("{{.users.confirm.title}}", input);
      var dbKey = "authentication.web";
      var input = content.createCheckbox(dbKey);
      input.checked = data[dbKey];
      if (data["defaultUser"] == true) {
        addRuntimeEventListener(input, "click", (event) => event.preventDefault());
      }
      content.appendRow("{{.users.web.title}}", input);
      var dbKey = "authentication.pms";
      var input = content.createCheckbox(dbKey);
      input.checked = data[dbKey];
      content.appendRow("{{.users.pms.title}}", input);
      var dbKey = "authentication.m3u";
      var input = content.createCheckbox(dbKey);
      input.checked = data[dbKey];
      content.appendRow("{{.users.m3u.title}}", input);
      var dbKey = "authentication.xml";
      var input = content.createCheckbox(dbKey);
      input.checked = data[dbKey];
      content.appendRow("{{.users.xml.title}}", input);
      var dbKey = "authentication.api";
      var input = content.createCheckbox(dbKey);
      input.checked = data[dbKey];
      content.appendRow("{{.users.api.title}}", input);
      var dbKey = "authentication.config";
      var input = content.createCheckbox(dbKey);
      input.checked = data[dbKey] == true;
      content.appendRow("{{.users.config.title}}", input);
      content.createInteraction();
      if (data["defaultUser"] != true && id != "-") {
        var input = content.createInput("button", "delete", "{{.button.delete}}");
        input.className = "delete";
        addRuntimeEventListener(input, "click", () => savePopupData(dataType, id, true, 0));
        content.addInteraction(input);
      }
      var input = content.createInput("button", "cancel", "{{.button.cancel}}");
      addRuntimeEventListener(input, "click", () => showElement("popup", false));
      content.addInteraction(input);
      var input = content.createInput("button", "save", "{{.button.save}}");
      addRuntimeEventListener(input, "click", () => savePopupData(dataType, id, false, void 0));
      content.addInteraction(input);
      break;
    case "mapping":
      content.createHeadline("{{.mainMenu.item.mapping}}");
      if (BULK_EDIT == true) {
        var dbKey = "x-channels-start";
        var input = content.createInput("text", dbKey, data[dbKey]);
        var channels = getAllSelectedChannels();
        var channel = SERVER["xepg"]["epgMapping"][channels[0]];
        if (typeof channel !== "undefined") {
          input.setAttribute("value", channel["x-channelID"]);
        }
        addRuntimeEventListener(input, "change", () => changeChannelNumbers(channels.toString()));
        content.appendRow("{{.mapping.channelGroupStart.title}}", input);
      }
      var dbKey = "x-active";
      var input = content.createCheckbox(dbKey);
      input.checked = data[dbKey];
      input.id = "active";
      addRuntimeEventListener(input, "change", () => toggleChannelStatus(id));
      content.appendRow("{{.mapping.active.title}}", input);
      var dbKey = "x-name";
      var input = content.createInput("text", dbKey, data[dbKey]);
      addRuntimeEventListener(input, "change", (event) => {
        event.currentTarget.className = "changed";
      });
      if (BULK_EDIT == true) {
        input.style.border = "solid 1px red";
        input.setAttribute("readonly", "true");
      }
      content.appendRow("{{.mapping.channelName.title}}", input);
      var mappingDescription = document.createElement("PRE");
      var mappingTvgID = document.createElement("SPAN");
      mappingTvgID.className = "text-danger";
      mappingTvgID.textContent = data["tvg-id"];
      mappingDescription.appendChild(mappingTvgID);
      mappingDescription.appendChild(document.createTextNode(" "));
      var mappingEPG = document.createElement("SPAN");
      mappingEPG.className = "text-primary";
      mappingEPG.textContent = "(" + data["x-epg"] + ")";
      mappingDescription.appendChild(mappingEPG);
      content.descriptionElement(mappingDescription);
      var dbKey = "x-description";
      var input = content.createInput("text", dbKey, data[dbKey]);
      input.setAttribute("placeholder", "{{.mapping.description.placeholder}}");
      addRuntimeEventListener(input, "change", (event) => {
        event.currentTarget.className = "changed";
      });
      content.appendRow("{{.mapping.description.title}}", input);
      if (data.hasOwnProperty("_uuid.key")) {
        if (data["_uuid.key"] != "") {
          var dbKey = "x-update-channel-name";
          var input = content.createCheckbox(dbKey);
          addRuntimeEventListener(input, "change", (event) => {
            event.currentTarget.className = "changed";
          });
          input.checked = data[dbKey];
          content.appendRow("{{.mapping.updateChannelName.title}}", input);
        }
      }
      var dbKey = "tvg-logo";
      var input = content.createInput("text", dbKey, data[dbKey]);
      addRuntimeEventListener(input, "change", (event) => {
        event.currentTarget.className = "changed";
      });
      input.setAttribute("id", "channel-icon");
      content.appendRow("{{.mapping.channelLogo.title}}", input);
      var dbKey = "x-update-channel-icon";
      var input = content.createCheckbox(dbKey);
      input.checked = data[dbKey];
      input.setAttribute("id", "update-icon");
      addRuntimeEventListener(input, "change", (event) => {
        var updateIcon = event.currentTarget;
        updateIcon.className = "changed";
        changeChannelLogo(id);
      });
      content.appendRow("{{.mapping.updateChannelLogo.title}}", input);
      var dbKey = "x-category";
      var text = ["-"];
      var values = [""];
      var epgCategories = SERVER["settings"]["epgCategories"];
      var categories = epgCategories.split("|");
      for (i = 0; i <= categories.length; i++) {
        var cat = categories[i];
        if (cat) {
          var cat_split = cat.split(":");
          text.push(cat_split[0]);
          values.push(cat_split[1]);
        }
      }
      var select = content.createSelect(text, values, data[dbKey], dbKey);
      addRuntimeEventListener(select, "change", (event) => {
        event.currentTarget.className = "changed";
      });
      content.appendRow("{{.mapping.epgCategory.title}}", select);
      var dbKey = "x-group-title";
      var input = content.createInput("text", dbKey, data[dbKey]);
      addRuntimeEventListener(input, "change", (event) => {
        event.currentTarget.className = "changed";
      });
      content.appendRow("{{.mapping.m3uGroupTitle.title}}", input);
      if (data["group-title"] != void 0) {
        content.description(data["group-title"]);
      }
      var dbKey = "x-xmltv-file";
      var xmlFile = data[dbKey];
      var xmltv = new XMLTVFile();
      var select = xmltv.getFiles(data[dbKey]);
      select.setAttribute("name", dbKey);
      select.setAttribute("id", "popup-xmltv");
      addRuntimeEventListener(select, "change", (event) => {
        var xmltvFile = event.currentTarget;
        xmltvFile.className = "changed";
        setXmltvChannel(id, xmltvFile);
      });
      content.appendRow("{{.mapping.xmltvFile.title}}", select);
      var file = data[dbKey];
      var dbKey = "x-mapping";
      var xmltv = new XMLTVFile();
      const currentXmlTvId = data[dbKey];
      const [xmlTvIdContainer, xmlTvIdInput, xmlTvIdDatalist] = xmltv.newXmlTvIdPicker(xmlFile, currentXmlTvId);
      xmlTvIdContainer.setAttribute("id", "xmltv-id-picker-container");
      xmlTvIdInput.setAttribute("list", "xmltv-id-picker-datalist");
      xmlTvIdInput.setAttribute("name", "x-mapping");
      xmlTvIdInput.setAttribute("id", "xmltv-id-picker-input");
      addRuntimeEventListener(xmlTvIdInput, "change", () => {
        xmlTvIdInput.className = "changed";
        checkXmltvChannel(id, xmlTvIdInput, xmlFile);
      });
      xmlTvIdDatalist.setAttribute("id", "xmltv-id-picker-datalist");
      content.appendRow("{{.mapping.xmltvChannel.title}}", xmlTvIdContainer);
      if (currentXmlTvId == "PPV") {
        var dbKey = "x-ppv-extra";
        var input = content.createInput("text", dbKey, data[dbKey]);
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        input.setAttribute("id", "ppv-extra");
        content.appendRow("{{.mapping.ppvextra.title}}", input);
      }
      var dbKey = "x-backup-channel-1";
      var xmltv = new XMLTVFile();
      const backup1XmlTvId = data[dbKey];
      const [xmlTvBackup1IdContainer, xmlTvBackup1IdInput, xmlTvBackup1IdDatalist] = xmltv.newM3uPicker(xmlFile, backup1XmlTvId);
      xmlTvBackup1IdContainer.setAttribute("id", "m3u-id-picker-container-1");
      xmlTvBackup1IdInput.setAttribute("list", "m3u-id-picker-datalist");
      xmlTvBackup1IdInput.setAttribute("name", dbKey);
      xmlTvBackup1IdInput.setAttribute("id", "backup-channel-1");
      addRuntimeEventListener(xmlTvBackup1IdInput, "change", () => {
        xmlTvBackup1IdInput.className = "changed";
        checkXmltvChannel(id, xmlTvBackup1IdInput, xmlFile);
      });
      xmlTvBackup1IdDatalist.setAttribute("id", "m3u-id-picker-datalist");
      content.appendRow("{{.mapping.backupChannel1.title}}", xmlTvBackup1IdContainer);
      var dbKey = "x-backup-channel-2";
      var xmltv = new XMLTVFile();
      const backup2XmlTvId = data[dbKey];
      const [xmlTvBackup2IdContainer, xmlTvBackup2IdInput, xmlTvBackup2IdDatalist] = xmltv.newM3uPicker(xmlFile, backup2XmlTvId);
      xmlTvBackup2IdContainer.setAttribute("id", "xmltv-id-picker-container-2");
      xmlTvBackup2IdInput.setAttribute("list", "m3u-id-picker-datalist");
      xmlTvBackup2IdInput.setAttribute("name", dbKey);
      xmlTvBackup2IdInput.setAttribute("id", "backup-channel-2");
      addRuntimeEventListener(xmlTvBackup2IdInput, "change", () => {
        xmlTvBackup2IdInput.className = "changed";
        checkXmltvChannel(id, xmlTvBackup2IdInput, xmlFile);
      });
      xmlTvBackup2IdDatalist.setAttribute("id", "m3u-id-picker-datalist");
      content.appendRow("{{.mapping.backupChannel2.title}}", xmlTvBackup2IdContainer);
      var dbKey = "x-backup-channel-3";
      var xmltv = new XMLTVFile();
      const backup3XmlTvId = data[dbKey];
      const [xmlTvBackup3IdContainer, xmlTvBackup3IdInput, xmlTvBackup3IdDatalist] = xmltv.newM3uPicker(xmlFile, backup3XmlTvId);
      xmlTvBackup3IdContainer.setAttribute("id", "xmltv-id-picker-container-3");
      xmlTvBackup3IdInput.setAttribute("list", "m3u-id-picker-datalist");
      xmlTvBackup3IdInput.setAttribute("name", dbKey);
      xmlTvBackup3IdInput.setAttribute("id", "backup-channel-3");
      addRuntimeEventListener(xmlTvBackup3IdInput, "change", () => {
        xmlTvBackup3IdInput.className = "changed";
        checkXmltvChannel(id, xmlTvBackup3IdInput, xmlFile);
      });
      xmlTvBackup3IdDatalist.setAttribute("id", "m3u-id-picker-datalist");
      content.appendRow("{{.mapping.backupChannel3.title}}", xmlTvBackup3IdContainer);
      content.createInteraction();
      var input = content.createInput("button", "cancel", "{{.button.probeChannel}}");
      addRuntimeEventListener(input, "click", () => probeChannel(data["url"]));
      content.addInteraction(input);
      var input = content.createInput("button", "cancel", "{{.button.uploadLogo}}");
      addRuntimeEventListener(input, "click", () => uploadLogo());
      content.addInteraction(input);
      var input = content.createInput("button", "cancel", "{{.button.cancel}}");
      addRuntimeEventListener(input, "click", () => showElement("popup", false));
      content.addInteraction(input);
      var ids = new Array();
      ids = getAllSelectedChannels();
      if (ids.length == 0) {
        ids.push(id);
      }
      var input = content.createInput("button", "save", "{{.button.done}}");
      addRuntimeEventListener(input, "click", () => donePopupData(dataType, ids.toString()));
      content.addInteraction(input);
      var probeDetails = document.createElement("div");
      probeDetails.id = "probeDetails";
      content.appendRow("{{.mapping.probeDetails.title}}", probeDetails);
      break;
    default:
      break;
  }
  enhanceSourcePopup(dataType);
  if (typeof enhanceFilterPopup == "function") {
    enhanceFilterPopup(dataType);
  }
  if (dataType == "users") {
    enhanceUsersPopup(id, data);
  }
  showPopUpElement("popup-custom");
}
var XMLTVFile = class {
  getFiles(set) {
    var fileIDs = getObjKeys(SERVER["xepg"]["xmltvMap"]);
    var values = new Array("-");
    var text = new Array("-");
    for (let i = 0; i < fileIDs.length; i++) {
      if (fileIDs[i] != "Threadfin Dummy") {
        values.push(getValueFromProviderFile(fileIDs[i], "xmltv", "file.threadfin"));
        text.push(getValueFromProviderFile(fileIDs[i], "xmltv", "name"));
      } else {
        values.push(fileIDs[i]);
        text.push(fileIDs[i]);
      }
    }
    var select = document.createElement("SELECT");
    for (let i = 0; i < text.length; i++) {
      var option = document.createElement("OPTION");
      option.setAttribute("value", values[i]);
      option.innerText = text[i];
      select.appendChild(option);
    }
    if (set != "") {
      select.value = set;
    }
    return select;
  }
  /**
  * @param xmlTvFile XML file path to get EPG from.
  * @param currentXmlTvId Current XMLTV ID to set initial input value to.
  * @returns Array of, sequentially:
  * 1) Container of the picker.
  * 2) Input field to type at and get choice from.
  * 3) Datalist containing every option.
  */
  newXmlTvIdPicker(xmlTvFile, currentXmlTvId) {
    const container = document.createElement("div");
    const input = document.createElement("input");
    input.setAttribute("type", "text");
    input.value = currentXmlTvId ? currentXmlTvId : "-";
    input.addEventListener("focus", (evt) => {
      const target = evt.target;
      target.value = target.value === "-" ? "" : target.value;
    });
    input.addEventListener("blur", setFallbackValue);
    input.addEventListener("change", setFallbackValue);
    function setFallbackValue(evt) {
      const target = evt.target;
      target.value = target.value ? target.value : "-";
    }
    container.appendChild(input);
    const datalist = document.createElement("datalist");
    const option = document.createElement("option");
    option.setAttribute("value", "-");
    option.innerText = "-";
    datalist.appendChild(option);
    const epg = SERVER["xepg"]["xmltvMap"][xmlTvFile];
    if (epg) {
      const programIds = getOwnObjProps(epg);
      programIds.forEach((programId) => {
        const program = epg[programId];
        if (program.hasOwnProperty("display-name")) {
          const option2 = document.createElement("option");
          option2.setAttribute("value", programId);
          option2.innerText = program["display-name"];
          datalist.appendChild(option2);
        } else {
          const option2 = document.createElement("option");
          option2.setAttribute("value", programId);
          option2.innerText = "-";
          datalist.appendChild(option2);
        }
      });
    }
    container.appendChild(datalist);
    return [container, input, datalist];
  }
  newM3uPicker(xmlTvFile, currentXmlTvId) {
    const container = document.createElement("div");
    const input = document.createElement("input");
    input.setAttribute("type", "text");
    input.value = currentXmlTvId ? currentXmlTvId : "-";
    input.addEventListener("focus", (evt) => {
      const target = evt.target;
      target.value = target.value === "-" ? "" : target.value;
    });
    input.addEventListener("blur", setFallbackValue);
    input.addEventListener("change", setFallbackValue);
    function setFallbackValue(evt) {
      const target = evt.target;
      target.value = target.value ? target.value : "-";
    }
    container.appendChild(input);
    const datalist = document.createElement("datalist");
    const option = document.createElement("option");
    option.setAttribute("value", "-");
    option.innerText = "-";
    datalist.appendChild(option);
    const m3u = SERVER["xepg"]["epgMapping"];
    if (m3u) {
      const programIds = getOwnObjProps(m3u);
      programIds.forEach((programId) => {
        const chanel = m3u[programId];
        if (chanel.hasOwnProperty("tvg-name")) {
          const option2 = document.createElement("option");
          option2.setAttribute("value", chanel["tvg-name"]);
          option2.innerText = programId;
          datalist.appendChild(option2);
        } else {
          const option2 = document.createElement("option");
          option2.setAttribute("value", chanel["tvg-name"]);
          option2.innerText = "-";
          datalist.appendChild(option2);
        }
      });
    }
    container.appendChild(datalist);
    return [container, input, datalist];
  }
  getPrograms(file, set, active) {
    var values = getObjKeys(SERVER["xepg"]["xmltvMap"][file]);
    var text = new Array();
    var displayName;
    var actives = getObjKeys(SERVER["data"]["StreamPreviewUI"]["activeStreams"]);
    var active_list = new Array();
    if (active == true) {
      for (let i = 0; i < actives.length; i++) {
        var names_split = SERVER["data"]["StreamPreviewUI"]["activeStreams"][actives[i]].split("[");
        displayName = names_split[0].trim();
        if (displayName != "") {
          var object = { "value": displayName, "display": displayName };
          active_list.push(object);
        }
      }
    } else {
      for (let i = 0; i < values.length; i++) {
        if (SERVER["xepg"]["xmltvMap"][file][values[i]].hasOwnProperty("display-name") == true) {
          displayName = SERVER["xepg"]["xmltvMap"][file][values[i]]["display-name"];
        } else {
          displayName = "-";
        }
        text[i] = displayName + " (" + values[i] + ")";
      }
    }
    text.unshift("-");
    values.unshift("-");
    var select = document.createElement("SELECT");
    for (let i = 0; i < text.length; i++) {
      var option = document.createElement("OPTION");
      option.setAttribute("value", values[i]);
      option.innerText = text[i];
      select.appendChild(option);
    }
    for (let i = 0; i < active_list.length; i++) {
      var option = document.createElement("OPTION");
      option.setAttribute("value", active_list[i]["value"]);
      option.innerText = active_list[i]["display"];
      select.appendChild(option);
    }
    if (set != "") {
      select.value = set;
    }
    if (select.value != set) {
      select.value = "-";
    }
    return select;
  }
};
function getValueFromProviderFile(file, fileType, key) {
  if (file == "Threadfin Dummy") {
    return file;
  }
  var fileID;
  var indicator = file.charAt(0);
  switch (indicator) {
    case "M":
      fileType = "m3u";
      fileID = file;
      break;
    case "H":
      fileType = "hdhr";
      fileID = file;
      break;
    case "X":
      fileType = "xmltv";
      fileID = file.substring(0, file.lastIndexOf("."));
      break;
  }
  if (SERVER["settings"]["files"][fileType].hasOwnProperty(fileID) == true) {
    var data = SERVER["settings"]["files"][fileType][fileID];
    return data[key];
  }
  return;
}
function setXmltvChannel(epgMapId, xmlTvFileSelect) {
  const xmlTv = new XMLTVFile();
  const newXmlTvFile = xmlTvFileSelect.value;
  const xmlTvIdPickerParent = document.getElementById("xmltv-id-picker-container").parentElement;
  xmlTvIdPickerParent.innerHTML = "";
  const tvgId = SERVER["xepg"]["epgMapping"][epgMapId]["tvg-id"];
  const [xmlTvIdContainer, xmlTvIdInput, xmlTvIdDatalist] = xmlTv.newXmlTvIdPicker(newXmlTvFile, tvgId);
  xmlTvIdContainer.setAttribute("id", "xmltv-id-picker-container");
  xmlTvIdInput.setAttribute("list", "xmltv-id-picker-datalist");
  xmlTvIdInput.setAttribute("name", "x-mapping");
  xmlTvIdInput.setAttribute("id", "xmltv-id-picker-input");
  addRuntimeEventListener(xmlTvIdInput, "change", () => {
    xmlTvIdInput.className = "changed";
    checkXmltvChannel(epgMapId, xmlTvIdInput, newXmlTvFile);
  });
  xmlTvIdInput.classList.add("changed");
  xmlTvIdDatalist.setAttribute("id", "xmltv-id-picker-datalist");
  xmlTvIdPickerParent.appendChild(xmlTvIdContainer);
  checkXmltvChannel(epgMapId, xmlTvIdInput.value, newXmlTvFile);
}
function checkXmltvChannel(id, element, xmlFile) {
  var value = element.value;
  var bool;
  var checkbox = document.getElementById("active");
  var channel = SERVER["xepg"]["epgMapping"][id];
  var updateLogo;
  if (value == "-") {
    bool = false;
  } else {
    bool = true;
  }
  checkbox.checked = bool;
  checkbox.className = "changed";
  console.log(xmlFile);
  if (xmlFile != "Threadfin Dummy" && bool == true) {
    console.log("ID", id);
    changeChannelLogo(id);
    return;
  }
  if (xmlFile == "Threadfin Dummy") {
    document.getElementById("update-icon").checked = false;
    document.getElementById("update-icon").className = "changed";
  }
  return;
}
function changeChannelLogo(epgMapId) {
  const channel = SERVER["xepg"]["epgMapping"][epgMapId];
  const xmlTvFileSelect = document.getElementById("popup-xmltv");
  const xmlTvFile = xmlTvFileSelect.options[xmlTvFileSelect.selectedIndex].value;
  const xmlTvIdInput = document.getElementById("xmltv-id-picker-input");
  const newXmlTvId = xmlTvIdInput.value;
  const updateLogo = !BULK_EDIT || document.getElementById("update-icon").checked;
  let logo;
  if (updateLogo == true && xmlTvFile != "Threadfin Dummy") {
    if (SERVER["xepg"]["xmltvMap"][xmlTvFile].hasOwnProperty(newXmlTvId)) {
      logo = SERVER["xepg"]["xmltvMap"][xmlTvFile][newXmlTvId]["icon"];
    } else {
      logo = channel["tvg-logo"];
    }
  }
}
function savePopupData(dataType, id, remove, option) {
  var filterPopupValid = typeof validateFilterPopup != "function" || validateFilterPopup(dataType);
  if (remove != true && option == 0 && (!validateSourcePopup(dataType) || !filterPopupValid)) {
    return;
  }
  if (dataType == "mapping") {
    showElement("loading", true);
    var data = new Object();
    console.log("Save mapping data");
    cmd = "saveEpgMapping";
    data["epgMapping"] = SERVER["xepg"]["epgMapping"];
    console.log("SEND TO SERVER");
    var server = new Server(cmd);
    server.request(data);
    delete UNDO["epgMapping"];
    showElement("loading", false);
    return;
  }
  console.log("Save popup data");
  var div = document.getElementById("popup-custom");
  var inputs = div.getElementsByTagName("TABLE")[0].getElementsByTagName("INPUT");
  var selects = div.getElementsByTagName("TABLE")[0].getElementsByTagName("SELECT");
  var input = new Object();
  var confirmMsg;
  for (let i = 0; i < selects.length; i++) {
    var name;
    name = selects[i].name;
    var value = selects[i].value;
    switch (name) {
      case "tuner":
        input[name] = parseInt(value);
        break;
      default:
        input[name] = value;
        break;
    }
  }
  for (let i = 0; i < inputs.length; i++) {
    switch (inputs[i].type) {
      case "checkbox":
        name = inputs[i].name;
        input[name] = inputs[i].checked;
        break;
      case "text":
      case "hidden":
      case "password":
        name = inputs[i].name;
        switch (name) {
          case "tuner":
            input[name] = parseInt(inputs[i].value);
            break;
          default:
            input[name] = inputs[i].value;
            break;
        }
        break;
    }
  }
  var data = new Object();
  var cmd;
  if (remove == true) {
    input["delete"] = true;
  }
  switch (dataType) {
    case "users":
      confirmMsg = "Delete this user?";
      var userRequest = buildUserRequest(id, input, remove == true);
      cmd = userRequest.cmd;
      data = userRequest.data;
      break;
    case "m3u":
      confirmMsg = "Delete this playlist?";
      switch (option) {
        // Popup: Save
        case 0:
          cmd = "saveFilesM3U";
          break;
        // Popup: Update
        case 1:
          cmd = "updateFileM3U";
          break;
      }
      data["files"] = new Object();
      data["files"][dataType] = new Object();
      data["files"][dataType][id] = input;
      break;
    case "hdhr":
      confirmMsg = "Delete this HDHomeRun tuner?";
      switch (option) {
        // Popup: Save
        case 0:
          cmd = "saveFilesHDHR";
          break;
        // Popup: Update
        case 1:
          cmd = "updateFileHDHR";
          break;
      }
      data["files"] = new Object();
      data["files"][dataType] = new Object();
      data["files"][dataType][id] = input;
      break;
    case "xmltv":
      confirmMsg = "Delete this XMLTV file?";
      switch (option) {
        // Popup: Save
        case 0:
          cmd = "saveFilesXMLTV";
          break;
        // Popup: Update
        case 1:
          cmd = "updateFileXMLTV";
          break;
      }
      data["files"] = new Object();
      data["files"][dataType] = new Object();
      data["files"][dataType][id] = input;
      break;
    case "filter":
      confirmMsg = "Delete this filter?";
      cmd = "saveFilter";
      data["filter"] = new Object();
      data["filter"][id] = input;
      break;
    default:
      return;
      break;
  }
  if (remove == true) {
    if (!confirm(confirmMsg)) {
      showElement("popup", false);
      return;
    }
  }
  showElement("loading", true);
  console.log("SEND TO SERVER");
  beginSourceRequest(dataType, id, remove, option);
  if (typeof beginFilterRequest == "function") {
    beginFilterRequest(dataType, id, remove);
  }
  var server = new Server(cmd);
  server.request(data);
  showElement("loading", false);
}
function donePopupData(dataType, idsStr) {
  var ids = idsStr.split(",");
  var div = document.getElementById("popup-custom");
  var inputs = div.getElementsByClassName("changed");
  ids.forEach((id) => {
    var input = new Object();
    input = SERVER["xepg"]["epgMapping"][id];
    console.log("INPUT: " + input);
    for (let i = 0; i < inputs.length; i++) {
      var name;
      var value;
      switch (inputs[i].tagName) {
        case "INPUT":
          switch (inputs[i].type) {
            case "checkbox":
              name = inputs[i].name;
              value = inputs[i].checked;
              input[name] = value;
              break;
            case "text":
              name = inputs[i].name;
              value = inputs[i].value;
              input[name] = value;
              break;
          }
          break;
        case "SELECT":
          name = inputs[i].name;
          value = inputs[i].value;
          input[name] = value;
          break;
      }
      switch (name) {
        case "tvg-logo":
          break;
        case "x-channel-start":
          document.getElementById(id).childNodes[3].firstChild.textContent = value;
          break;
        case "x-name":
          document.getElementById(id).childNodes[3].firstChild.textContent = value;
          break;
        case "x-category":
          var color = "white";
          var catColorSettings = SERVER["settings"]["epgCategoriesColors"];
          var colors_split = catColorSettings.split("|");
          for (var ii = 0; ii < colors_split.length; ii++) {
            var catsColor_split = colors_split[ii].split(":");
            if (catsColor_split[0] == value) {
              color = catsColor_split[1];
            }
          }
          document.getElementById(id).childNodes[3].firstChild.style.borderColor = color;
          break;
        case "x-group-title":
          document.getElementById(id).childNodes[5].firstChild.textContent = value;
          break;
        case "x-xmltv-file":
          if (value != "Threadfin Dummy" && value != "-") {
            value = getValueFromProviderFile(value, "xmltv", "name");
          }
          if (value == "-") {
            input["x-active"] = false;
          }
          document.getElementById(id).childNodes[6].firstChild.textContent = value;
          break;
        case "x-mapping":
          if (value == "-") {
            input["x-active"] = false;
          }
          document.getElementById(id).childNodes[7].firstChild.textContent = value;
          break;
        case "x-backup-channel":
          document.getElementById(id).childNodes[7].firstChild.textContent = value;
          break;
        case "x-hide-channel":
          document.getElementById(id).childNodes[7].firstChild.textContent = value;
          break;
        default:
      }
      createSearchObj();
      searchInMapping();
    }
    if (input["x-active"] == false) {
      document.getElementById(id).className = "notActiveEPG";
    } else {
      document.getElementById(id).className = "activeEPG";
    }
    console.log(input["tvg-logo"]);
    document.getElementById(id).childNodes[2].firstChild.firstChild.setAttribute("src", input["tvg-logo"]);
  });
  showElement("popup", false);
  return;
}
function showPreview(element) {
  var div = document.getElementById("myStreamsBox");
  switch (element) {
    case false:
      div.className = "notVisible";
      return;
      break;
  }
  var streams = ["activeStreams", "inactiveStreams"];
  streams.forEach((preview) => {
    var table = document.getElementById(preview);
    table.innerHTML = "";
    var obj = SERVER["data"]["StreamPreviewUI"][preview];
    var caption = document.createElement("CAPTION");
    var result = preview.replace(/([A-Z])/g, " $1");
    var finalResult = result.charAt(0).toUpperCase() + result.slice(1);
    caption.textContent = finalResult;
    table.appendChild(caption);
    var tbody = document.createElement("TBODY");
    table.appendChild(tbody);
    obj.slice(0, 1e3).forEach((channel) => {
      var tr = document.createElement("TR");
      var tdKey = document.createElement("TD");
      var tdVal = document.createElement("TD");
      tdKey.className = "tdKey";
      tdVal.className = "tdVal";
      switch (preview) {
        case "activeStreams":
          tdKey.innerText = "Channel: (+)";
          break;
        case "inactiveStreams":
          tdKey.innerText = "Channel: (-)";
          break;
      }
      tdVal.innerText = channel;
      tr.appendChild(tdKey);
      tr.appendChild(tdVal);
      tbody.appendChild(tr);
      table.appendChild(tr);
    });
  });
  div.className = "visible";
  return;
}
document.addEventListener("DOMContentLoaded", function() {
  var backToTopButton = document.createElement("button");
  backToTopButton.id = "back-to-top";
  backToTopButton.innerText = "Back to Top";
  backToTopButton.onclick = function() {
    window.scrollTo({ top: 0, behavior: "smooth" });
  };
  document.body.appendChild(backToTopButton);
  window.addEventListener("scroll", function() {
    if (window.scrollY > 200) {
      backToTopButton.style.display = "block";
    } else {
      backToTopButton.style.display = "none";
    }
  });
});
var USER_PERMISSION_DEFINITIONS = [
  { key: "authentication.web", label: "WEB", description: "Sign in to the Threadfin web interface and manage this user's access." },
  { key: "authentication.pms", label: "PMS", description: "Authenticate DVR discovery and access used by Plex, Emby, or Jellyfin." },
  { key: "authentication.m3u", label: "M3U", description: "Authenticate requests for the generated M3U playlist." },
  { key: "authentication.xml", label: "XML", description: "Authenticate requests for the generated XMLTV guide." },
  { key: "authentication.api", label: "API", description: "Authenticate access to Threadfin's API commands." },
  { key: "authentication.config", label: "CONFIG", description: "Authorize the typed configuration workflow, including source fetches from permitted private-LAN destinations." }
];
function userActions(data) {
  var defaultUser = Boolean(data && data.defaultUser);
  return { canDelete: !defaultUser, webLocked: defaultUser };
}
function userRowValues(record) {
  var data = record && record.data && typeof record.data == "object" ? record.data : {};
  var row = { username: String(data.username || "") };
  USER_PERMISSION_DEFINITIONS.forEach((permission) => {
    row[permission.key] = data[permission.key] === true;
  });
  row.defaultUser = data.defaultUser === true;
  return row;
}
function buildUserRequest(id, input, remove) {
  var values = {};
  Object.keys(input || {}).forEach((key) => {
    values[key] = input[key];
  });
  if (remove) {
    values.delete = true;
  }
  if (id == "-") {
    return { cmd: "saveNewUser", data: { userData: values } };
  }
  var users = {};
  users[id] = values;
  return { cmd: "saveUserData", data: { userData: users } };
}
function enhanceUsersPopup(id, data) {
  var popup = document.getElementById("popup-custom");
  if (!popup) {
    return;
  }
  var username = popup.querySelector('[name="username"]');
  var password = popup.querySelector('[name="password"]');
  var confirmPassword = popup.querySelector('[name="confirm"]');
  if (username) {
    username.autocomplete = "username";
  }
  if (password) {
    password.autocomplete = "new-password";
  }
  if (confirmPassword) {
    confirmPassword.autocomplete = "new-password";
  }
  var actions = userActions(data);
  var web = popup.querySelector('[name="authentication.web"]');
  if (web && actions.webLocked) {
    web.disabled = true;
    web.setAttribute("aria-describedby", "default-user-web-note");
    var note = document.createElement("p");
    note.id = "default-user-web-note";
    note.className = "tf-field-note";
    note.textContent = "WEB access is required for the default user.";
    web.parentElement.appendChild(note);
  }
}
function appendUserPermissionDetails(identity, values) {
  var details = document.createElement("details");
  details.className = "tf-user-permission-details";
  var summary = document.createElement("summary");
  summary.textContent = "Permissions";
  details.appendChild(summary);
  var list = document.createElement("dl");
  USER_PERMISSION_DEFINITIONS.forEach((permission) => {
    var term = document.createElement("dt");
    term.textContent = permission.label;
    var description = document.createElement("dd");
    description.textContent = (values[permission.key] ? "Allowed. " : "Denied. ") + permission.description;
    list.appendChild(term);
    list.appendChild(description);
  });
  details.appendChild(list);
  identity.appendChild(details);
}
function renderUsersPage(doc) {
  doc.innerHTML = "";
  var root = document.createElement("div");
  root.className = "tf-users";
  var header = document.createElement("header");
  header.className = "tf-admin-header";
  var headingGroup = document.createElement("div");
  var heading = document.createElement("h1");
  heading.textContent = "Users";
  var purpose = document.createElement("p");
  purpose.textContent = "Control which Threadfin interfaces and generated endpoints each account may use. Passwords are available only while creating or editing an account.";
  headingGroup.appendChild(heading);
  headingGroup.appendChild(purpose);
  var add = document.createElement("button");
  add.type = "button";
  add.className = "tf-primary-action";
  add.textContent = "New user";
  add.addEventListener("click", () => openPopUp("users", void 0));
  header.appendChild(headingGroup);
  header.appendChild(add);
  root.appendChild(header);
  var explanations = document.createElement("dl");
  explanations.className = "tf-permission-guide";
  USER_PERMISSION_DEFINITIONS.forEach((permission) => {
    var term = document.createElement("dt");
    term.textContent = permission.label;
    var description = document.createElement("dd");
    description.textContent = permission.description;
    explanations.appendChild(term);
    explanations.appendChild(description);
  });
  root.appendChild(explanations);
  var scroll = document.createElement("div");
  scroll.className = "tf-admin-table-scroll";
  var table = document.createElement("table");
  table.className = "tf-users-table";
  var caption = document.createElement("caption");
  caption.textContent = "Threadfin users and endpoint permissions";
  table.appendChild(caption);
  var head = document.createElement("thead");
  var headRow = document.createElement("tr");
  ["Username"].concat(USER_PERMISSION_DEFINITIONS.map((permission) => permission.label), ["Actions"]).forEach((label) => {
    var th = document.createElement("th");
    th.scope = "col";
    th.textContent = label;
    headRow.appendChild(th);
  });
  head.appendChild(headRow);
  table.appendChild(head);
  var body = document.createElement("tbody");
  var users = SERVER && SERVER["users"] && typeof SERVER["users"] == "object" ? SERVER["users"] : {};
  Object.keys(users).sort((left, right) => userRowValues(users[left]).username.localeCompare(userRowValues(users[right]).username)).forEach((userID) => {
    var values = userRowValues(users[userID]);
    var tr = document.createElement("tr");
    var username = document.createElement("th");
    username.scope = "row";
    username.className = "tf-user-identity";
    username.textContent = values.username || "Unnamed user";
    if (values.defaultUser) {
      var defaultLabel = document.createElement("span");
      defaultLabel.className = "tf-user-default";
      defaultLabel.textContent = "Default user";
      username.appendChild(defaultLabel);
    }
    appendUserPermissionDetails(username, values);
    tr.appendChild(username);
    USER_PERMISSION_DEFINITIONS.forEach((permission) => {
      var td = document.createElement("td");
      td.className = "tf-user-permission-cell";
      td.textContent = values[permission.key] ? "Allowed" : "Denied";
      td.setAttribute("data-allowed", values[permission.key] ? "true" : "false");
      tr.appendChild(td);
    });
    var actions = document.createElement("td");
    actions.className = "tf-user-actions";
    var edit = document.createElement("button");
    edit.type = "button";
    edit.id = userID;
    edit.textContent = "Edit";
    edit.addEventListener("click", () => openPopUp("users", edit));
    actions.appendChild(edit);
    tr.appendChild(actions);
    body.appendChild(tr);
  });
  table.appendChild(body);
  scroll.appendChild(table);
  root.appendChild(scroll);
  doc.appendChild(root);
}
var mappingWorkspaceState;
var mappingCurrentQuery = { sort: "number" };
var mappingPendingNavigation;
var mappingAfterSave;
var mappingProbeFeedback = "";
var mappingGuardDialog;
var mappingGuardReturnFocus;
var mappingBeforeUnloadBound = false;
var mappingFilterChips;
var mappingEditorDrawer;
var mappingEPGCombobox;
var mappingEditorBreakpoint;
var mappingEditorResize;
function initializeMappingWorkspace(server, force = false) {
  var authoritative = mappingAuthoritativeMap(server);
  if (force || !mappingWorkspaceState || mappingDirtyIDs(mappingWorkspaceState).length == 0 && mappingComparable(mappingWorkspaceState.baseline) != mappingComparable(authoritative)) {
    mappingWorkspaceState = createMappingWorkspaceState(server);
    mappingCurrentQuery.segment = mappingWorkspaceState.segment;
  } else {
    mappingWorkspaceState.xmltvMap = mappingDeepClone(mappingXMLTVMap(server));
  }
  bindMappingBeforeUnload();
  return mappingWorkspaceState;
}
function mappingHasDirtyDraft() {
  return !!mappingWorkspaceState && mappingDirtyIDs(mappingWorkspaceState).length > 0;
}
function saveMappingDraft(afterSave) {
  if (!mappingWorkspaceState || !mappingHasDirtyDraft() || mappingWorkspaceState.saveState == "pending" || mappingWorkspaceState.saveState == "ambiguous") {
    return;
  }
  mappingWorkspaceState.saveState = "pending";
  mappingWorkspaceState.feedback = "Saving mapping and rebuilding outputs";
  mappingAfterSave = afterSave;
  updateMappingGuardButtons();
  renderMappingWorkspaceIfOpen();
  var server = new Server("saveEpgMapping");
  server.request({ epgMapping: mappingDeepClone(mappingWorkspaceState.draft) });
}
function mappingResponseAuthoritative(response) {
  var root = mappingRecord(response);
  var xepg = mappingRecord(root.xepg);
  var present = Object.prototype.hasOwnProperty.call(xepg, "epgMapping");
  return { present, mapping: mappingRecord(xepg.epgMapping), xmltvMap: mappingRecord(xepg.xmltvMap) };
}
function publishMappingAuthoritativeResponse(authoritative) {
  SERVER["xepg"] = Object.assign({}, mappingRecord(SERVER["xepg"]), {
    epgMapping: mappingDeepClone(authoritative.mapping),
    xmltvMap: mappingDeepClone(authoritative.xmltvMap)
  });
}
function completeMappingRequest(command, data, response, failureKind) {
  if (command == "probeChannel") {
    completeMappingProbe(response);
    return;
  }
  if (command == "getServerConfig" && mappingWorkspaceState && mappingWorkspaceState.saveState == "ambiguous") {
    var refetched = mappingResponseAuthoritative(response);
    if (mappingRecord(response).status === true && refetched.present) {
      publishMappingAuthoritativeResponse(refetched);
      mappingWorkspaceState.xmltvMap = mappingDeepClone(refetched.xmltvMap);
      mappingReconcileAuthoritative(mappingWorkspaceState, refetched.mapping, mappingRecord(data).epgMapping, false);
      mappingWorkspaceState.saveState = "idle";
      mappingWorkspaceState.feedback = mappingHasDirtyDraft() ? "Authoritative mapping reloaded; review the retained draft before saving again" : "Authoritative mapping reloaded";
    } else {
      mappingWorkspaceState.feedback = "Save result remains unknown; authoritative mapping could not be reloaded";
    }
    renderMappingWorkspaceIfOpen();
    return;
  }
  if (command != "saveEpgMapping" || !mappingWorkspaceState) {
    return;
  }
  if (failureKind == "busy") {
    mappingWorkspaceState.saveState = "idle";
    mappingWorkspaceState.feedback = "Another request is active; the mapping draft was retained";
    renderMappingWorkspaceIfOpen();
    return;
  }
  if (failureKind == "transport") {
    mappingWorkspaceState.saveState = "ambiguous";
    mappingWorkspaceState.feedback = "Connection lost after save; checking the authoritative mapping before allowing another save";
    renderMappingWorkspaceIfOpen();
    new Server("getServerConfig").request(new Object());
    return;
  }
  var root = mappingRecord(response);
  var authoritative = mappingResponseAuthoritative(root);
  if (!authoritative.present) {
    mappingWorkspaceState.saveState = "idle";
    mappingWorkspaceState.feedback = "The server response did not include the authoritative mapping; the draft was retained";
    renderMappingWorkspaceIfOpen();
    return;
  }
  publishMappingAuthoritativeResponse(authoritative);
  mappingWorkspaceState.xmltvMap = mappingDeepClone(authoritative.xmltvMap);
  var reconciliation = mappingReconcileAuthoritative(mappingWorkspaceState, authoritative.mapping, mappingRecord(data).epgMapping, root.status === true);
  mappingWorkspaceState.saveState = "idle";
  if (root.status === true && root.mappingSaveResult == "outputsRebuilt") {
    mappingWorkspaceState.feedback = "Mapping saved; outputs rebuilt";
  } else if (root.status === true && root.mappingSaveResult == "outputRebuildRequested") {
    mappingWorkspaceState.feedback = "Mapping saved; output rebuild requested";
  } else if (root.status === true) {
    mappingWorkspaceState.feedback = "Mapping saved; output rebuild state was not confirmed";
  } else if (reconciliation.persisted) {
    mappingWorkspaceState.feedback = "Mapping saved, but output rebuilding failed" + (mappingString(root.err) ? ": " + mappingString(root.err) : "");
  } else {
    mappingWorkspaceState.feedback = mappingString(root.err) || "Mapping was not saved; the draft was retained";
  }
  var continuation = reconciliation.persisted ? mappingAfterSave : void 0;
  mappingAfterSave = void 0;
  renderMappingWorkspaceIfOpen();
  if (continuation) {
    closeMappingGuard();
    continuation();
  }
}
function requestMappingProbe(url) {
  mappingProbeFeedback = "Probing channel\u2026";
  renderMappingProbeFeedback();
  new Server("probeChannel").request({ probeUrl: url });
}
function completeMappingProbe(response) {
  var root = mappingRecord(response);
  if (root.status !== true) {
    mappingProbeFeedback = mappingString(root.err) || "Probe failed";
  } else {
    var info = mappingRecord(root.probeInfo);
    var values = [];
    if (info.resolution)
      values.push("Resolution: " + mappingString(info.resolution));
    if (info.frameRate)
      values.push("Frame rate: " + mappingString(info.frameRate) + " FPS");
    if (info.audioChannel)
      values.push("Audio: " + mappingString(info.audioChannel));
    mappingProbeFeedback = values.length ? values.join(" \xB7 ") : "Details unavailable";
  }
  renderMappingProbeFeedback();
}
function renderMappingProbeFeedback() {
  var status = document.getElementById("mapping-probe-status");
  if (status) {
    status.textContent = mappingProbeFeedback;
  }
}
function renderMappingWorkspaceIfOpen() {
  if (typeof document == "undefined")
    return;
  var host = document.getElementById("content");
  if (host && (typeof currentDestination == "undefined" || currentDestination == "mapping")) {
    var focus = captureMappingWorkspaceFocus(host);
    renderMappingPage(host);
    restoreMappingWorkspaceFocus(focus);
  }
  updateMappingGuardButtons();
  updateMappingMutationControls();
}
function captureMappingWorkspaceFocus(host) {
  var active = document.activeElement;
  if (!active || !host.contains(active))
    return void 0;
  var key = active.getAttribute("data-mapping-focus");
  if (!key)
    return void 0;
  var snapshot = { key };
  if (typeof active.selectionStart == "number") {
    snapshot.selectionStart = active.selectionStart;
    snapshot.selectionEnd = active.selectionEnd;
  }
  return snapshot;
}
function restoreMappingWorkspaceFocus(snapshot) {
  if (!snapshot)
    return;
  var controls = document.querySelectorAll("[data-mapping-focus]");
  var target;
  Array.prototype.some.call(controls, function(control) {
    if (control.getAttribute("data-mapping-focus") == snapshot.key) {
      target = control;
      return true;
    }
    return false;
  });
  if (!target)
    return;
  target.focus();
  if (snapshot.selectionStart !== void 0 && typeof target.setSelectionRange == "function") {
    try {
      target.setSelectionRange(snapshot.selectionStart, snapshot.selectionEnd);
    } catch (_error) {
    }
  }
}
function mappingRequestNavigation(continuation, invoker) {
  if (!mappingHasDirtyDraft()) {
    return true;
  }
  var candidate = arguments.length > 1 ? invoker : document.activeElement;
  mappingGuardReturnFocus = candidate;
  mappingPendingNavigation = continuation;
  openMappingGuard();
  return false;
}
function openMappingGuard() {
  if (mappingGuardDialog && mappingGuardDialog.parentElement) {
    updateMappingGuardButtons();
    return;
  }
  var guard = document.createElement("dialog");
  guard.className = "dialog tf-mapping-guard";
  guard.setAttribute("role", "dialog");
  guard.setAttribute("aria-modal", "true");
  guard.setAttribute("aria-labelledby", "mapping-guard-title");
  var panel = document.createElement("div");
  panel.className = "dialog-body tf-mapping-guard-panel";
  var title = document.createElement("h2");
  title.id = "mapping-guard-title";
  title.textContent = "Unsaved mapping draft";
  var copy = document.createElement("p");
  copy.textContent = "Save or discard staged Mapping changes before leaving.";
  var actions = document.createElement("div");
  actions.className = "tf-mapping-guard-actions";
  var save = mappingButton("Save mapping", function() {
    saveMappingDraft(mappingPendingNavigation);
  });
  save.setAttribute("data-mapping-guard", "save");
  var discard = mappingButton("Discard draft", function() {
    if (mappingWorkspaceState && mappingWorkspaceState.saveState == "idle") {
      mappingRevertDraft(mappingWorkspaceState);
      var continuation = mappingPendingNavigation;
      closeMappingGuard();
      renderMappingWorkspaceIfOpen();
      if (continuation)
        continuation();
    }
  });
  discard.setAttribute("data-mapping-guard", "discard");
  var stay = mappingButton("Stay", stayOnMapping);
  stay.setAttribute("data-mapping-guard", "stay");
  actions.appendChild(save);
  actions.appendChild(discard);
  actions.appendChild(stay);
  panel.appendChild(title);
  panel.appendChild(copy);
  panel.appendChild(actions);
  guard.appendChild(panel);
  mappingOverlayHost().appendChild(guard);
  mappingGuardDialog = guard;
  guard.addEventListener("cancel.bs.dialog", function() {
    mappingAfterSave = void 0;
  });
  guard.addEventListener("hidden.bs.dialog", function() {
    if (mappingGuardDialog == guard) {
      stayOnMapping();
    }
  });
  updateMappingGuardButtons();
  var dialog = new Dialog(guard);
  dialog.show().then(function() {
    stay.focus();
  });
}
function updateMappingGuardButtons() {
  if (!mappingGuardDialog || !mappingWorkspaceState)
    return;
  var disabled = mappingWorkspaceState.saveState == "pending" || mappingWorkspaceState.saveState == "ambiguous";
  var controls = mappingGuardDialog.querySelectorAll("button");
  Array.prototype.forEach.call(controls, function(button) {
    button.disabled = button.getAttribute("data-mapping-guard") == "stay" ? false : disabled;
  });
}
function stayOnMapping() {
  var returnFocus = mappingGuardReturnFocus;
  mappingAfterSave = void 0;
  closeMappingGuard();
  restoreMappingGuardFocus(returnFocus);
}
function closeMappingGuard() {
  if (mappingGuardDialog) {
    var instance = Dialog.getInstance(mappingGuardDialog);
    if (instance)
      instance.dispose();
    if (mappingGuardDialog.parentElement)
      mappingGuardDialog.parentElement.removeChild(mappingGuardDialog);
  }
  mappingGuardDialog = void 0;
  mappingPendingNavigation = void 0;
  mappingGuardReturnFocus = void 0;
}
function mappingGuardFocusIsAvailable(target) {
  return !!target && typeof target.focus == "function" && document.contains(target) && !target.disabled && target.getAttribute("aria-disabled") != "true" && !target.hidden;
}
function restoreMappingGuardFocus(returnFocus) {
  if (mappingGuardFocusIsAvailable(returnFocus)) {
    returnFocus.focus();
    return;
  }
  var fallback = document.getElementById("mapping-heading");
  if (!mappingGuardFocusIsAvailable(fallback)) {
    fallback = document.querySelector(".tf-mapping");
  }
  if (mappingGuardFocusIsAvailable(fallback)) {
    if (fallback.getAttribute("tabindex") == null)
      fallback.setAttribute("tabindex", "-1");
    fallback.focus();
  }
}
function bindMappingBeforeUnload() {
  if (mappingBeforeUnloadBound || typeof window == "undefined")
    return;
  mappingBeforeUnloadBound = true;
  window.addEventListener("beforeunload", function(event) {
    if (mappingHasDirtyDraft()) {
      event.preventDefault();
      event.returnValue = "";
    }
  });
}
function mappingOverlayHost() {
  return document.querySelector(".tf-app") || document.body;
}
function renderMappingPage(host) {
  if (mappingFilterChips) {
    mappingFilterChips.dispose();
    mappingFilterChips = void 0;
  }
  host.innerHTML = "";
  var page = document.createElement("div");
  page.className = "tf-mapping";
  var pmsMode = mappingString(mappingRecord(mappingRecord(SERVER).settings).epgSource).toUpperCase() == "PMS";
  if (!pmsMode) {
    initializeMappingWorkspace(SERVER);
  }
  page.appendChild(renderMappingHeader(!pmsMode));
  if (pmsMode) {
    var note = document.createElement("section");
    note.className = "tf-mapping-empty tf-mapping-mode-note";
    note.setAttribute("role", "note");
    var title = document.createElement("h2");
    title.textContent = "Mapping requires XEPG";
    var copy = document.createElement("p");
    copy.textContent = "PMS mode leaves guide data management to the connected client. Switch the EPG Source to XEPG to map channels and generate M3U/XMLTV guide outputs in Threadfin.";
    var settings = mappingButton("Open EPG Source settings", function() {
      openDestination("settings", true, settings);
    });
    note.appendChild(title);
    note.appendChild(copy);
    note.appendChild(settings);
    page.appendChild(note);
    host.appendChild(page);
    return;
  }
  page.appendChild(renderMappingFilters());
  var rows = mappingVisibleRows(mappingWorkspaceState, mappingCurrentQuery);
  page.appendChild(renderMappingCounts(rows));
  if (mappingWorkspaceState.selected.size > 0)
    page.appendChild(renderMappingBulkBar(rows));
  page.appendChild(renderMappingTable(rows));
  page.appendChild(renderMappingBackupOptions());
  page.appendChild(renderMappingSaveBar());
  host.appendChild(page);
  initializeMappingFilterChips();
}
function renderMappingHeader(showViews = true) {
  var header = document.createElement("header");
  header.className = "tf-mapping-header";
  var group = document.createElement("div");
  var title = document.createElement("h1");
  title.id = "mapping-heading";
  title.setAttribute("tabindex", "-1");
  title.textContent = "Mapping";
  var purpose = document.createElement("p");
  purpose.textContent = "Assign guide data, channel metadata, numbering, and backup streams before saving the complete lineup.";
  group.appendChild(title);
  group.appendChild(purpose);
  header.appendChild(group);
  if (!showViews) {
    return header;
  }
  var segments = document.createElement("div");
  segments.className = "tf-mapping-segments";
  segments.setAttribute("role", "group");
  segments.setAttribute("aria-label", "Mapping view");
  [["attention", "Needs attention"], ["active", "Active"], ["inactive", "Inactive"]].forEach(function(entry) {
    var segment = entry[0];
    var button = mappingButton(entry[1], function() {
      mappingWorkspaceState.segment = segment;
      mappingCurrentQuery.segment = segment;
      renderMappingWorkspaceIfOpen();
    });
    button.setAttribute("data-mapping-focus", "segment-" + segment);
    button.setAttribute("aria-pressed", String((mappingCurrentQuery.segment || mappingWorkspaceState.segment) == segment));
    segments.appendChild(button);
  });
  header.appendChild(segments);
  return header;
}
function renderMappingFilters() {
  var region = document.createElement("section");
  region.className = "tf-mapping-filters";
  region.setAttribute("aria-label", "Mapping search and filters");
  region.appendChild(mappingFilterInput("Search mapping", mappingString(mappingCurrentQuery.search), function(value) {
    mappingCurrentQuery.search = value;
  }));
  region.appendChild(mappingFilterSelect("Playlist", "playlist", mappingDistinctValues("_file.m3u.id"), mappingString(mappingCurrentQuery.playlist)));
  region.appendChild(mappingFilterSelect("Group", "group", mappingDistinctGroups(), mappingString(mappingCurrentQuery.group)));
  region.appendChild(mappingFilterSelect("XMLTV source", "xmltv", mappingDistinctValues("x-xmltv-file"), mappingString(mappingCurrentQuery.xmltv)));
  region.appendChild(mappingFilterSelect("Activation", "activation", [{ value: "active", label: "Active" }, { value: "inactive", label: "Inactive" }], mappingString(mappingCurrentQuery.activation)));
  region.appendChild(mappingFilterSelect("Attention reason", "reason", [
    { value: "missing", label: "Missing EPG assignment" },
    { value: "invalid", label: "Invalid EPG assignment" },
    { value: "hidden", label: "Hidden from outputs" },
    { value: "inactive", label: "Inactive" }
  ], mappingString(mappingCurrentQuery.reason)));
  var sortField = document.createElement("label");
  sortField.className = "form-field";
  sortField.textContent = "Sort by";
  var sortSelect = document.createElement("select");
  sortSelect.setAttribute("data-mapping-focus", "sort-order");
  [["number", "Channel"], ["name", "Name"], ["playlist", "Playlist"], ["group", "Group"], ["xmltv", "XMLTV / guide"]].forEach(function(entry) {
    ;
    ["asc", "desc"].forEach(function(direction) {
      var option = document.createElement("option");
      option.value = entry[0] + ":" + direction;
      option.textContent = entry[1] + (direction == "asc" ? " (ascending)" : " (descending)");
      sortSelect.appendChild(option);
    });
  });
  sortSelect.value = (mappingCurrentQuery.sort || "number") + (mappingCurrentQuery.descending ? ":desc" : ":asc");
  sortSelect.addEventListener("change", function() {
    var order = sortSelect.value.split(":");
    mappingCurrentQuery.sort = order[0];
    mappingCurrentQuery.descending = order[1] == "desc";
    renderMappingWorkspaceIfOpen();
  });
  sortField.appendChild(sortSelect);
  region.appendChild(sortField);
  region.appendChild(renderMappingFilterChips());
  return region;
}
function renderMappingFilterChips() {
  var chips = document.createElement("div");
  chips.id = "mapping-filter-chips";
  chips.className = "chips tf-mapping-filter-chips";
  chips.setAttribute("aria-label", "Active mapping filters");
  [["search", "Search"], ["playlist", "Playlist"], ["group", "Group"], ["xmltv", "XMLTV source"], ["activation", "Activation"], ["reason", "Attention reason"]].forEach(function(entry) {
    var value = mappingString(mappingCurrentQuery[entry[0]]);
    if (!value)
      return;
    var chip = document.createElement("span");
    chip.className = "chip";
    chip.setAttribute("data-mapping-filter", entry[0]);
    chip.setAttribute("data-bs-chip-value", entry[0]);
    chip.textContent = entry[1] + ": " + mappingProviderLabel(entry[0] == "playlist" ? "_file.m3u.id" : entry[0] == "xmltv" ? "x-xmltv-file" : entry[0], value);
    chips.appendChild(chip);
  });
  return chips;
}
function initializeMappingFilterChips() {
  var region = document.getElementById("mapping-filter-chips");
  if (!region)
    return;
  mappingFilterChips = new Chips(region, { createOnBlur: false, separator: null });
  var input = region.querySelector("input");
  if (input) {
    input.hidden = true;
    input.disabled = true;
  }
  region.addEventListener("add.bs.chips", function(event) {
    event.preventDefault();
  });
  region.addEventListener("change.bs.chips", function() {
    var remaining = mappingFilterChips.getValues();
    ["search", "playlist", "group", "xmltv", "activation", "reason"].forEach(function(key) {
      if (remaining.indexOf(key) == -1)
        mappingCurrentQuery[key] = "";
    });
    renderMappingWorkspaceIfOpen();
    var search = document.querySelector('[data-mapping-focus="filter-search"]');
    if (search)
      search.focus();
  });
}
function mappingFilterInput(label, value, update) {
  var wrapper = document.createElement("label");
  wrapper.textContent = label;
  var input = document.createElement("input");
  input.type = "search";
  input.value = value;
  input.setAttribute("aria-label", label);
  input.setAttribute("data-mapping-focus", "filter-search");
  input.addEventListener("input", function() {
    update(input.value);
    renderMappingWorkspaceIfOpen();
  });
  wrapper.appendChild(input);
  return wrapper;
}
function mappingFilterSelect(label, key, options, selected) {
  var wrapper = document.createElement("label");
  wrapper.textContent = label;
  var select = document.createElement("select");
  select.setAttribute("aria-label", label);
  select.setAttribute("data-mapping-focus", "filter-" + key);
  var all = document.createElement("option");
  all.value = "";
  all.textContent = "All";
  select.appendChild(all);
  options.forEach(function(entry) {
    var option = document.createElement("option");
    option.value = entry.value;
    option.textContent = entry.label;
    select.appendChild(option);
  });
  select.value = selected;
  select.addEventListener("change", function() {
    mappingCurrentQuery[key] = select.value;
    renderMappingWorkspaceIfOpen();
  });
  wrapper.appendChild(select);
  return wrapper;
}
function mappingDistinctValues(key) {
  var values = /* @__PURE__ */ new Set();
  Object.keys(mappingWorkspaceState.draft).forEach(function(id) {
    var value = mappingString(mappingWorkspaceState.draft[id][key]);
    if (value)
      values.add(value);
  });
  return Array.from(values).sort().map(function(value) {
    return { value, label: mappingProviderLabel(key, value) };
  });
}
function mappingDistinctGroups() {
  var values = /* @__PURE__ */ new Set();
  Object.keys(mappingWorkspaceState.draft).forEach(function(id) {
    var channel = mappingWorkspaceState.draft[id];
    var value = mappingString(channel["x-group-title"] || channel["group-title"]);
    if (value)
      values.add(value);
  });
  return Array.from(values).sort().map(function(value) {
    return { value, label: value };
  });
}
function mappingProviderLabel(key, value) {
  if (value == "-" || value == "Threadfin Dummy")
    return value;
  var files = mappingRecord(mappingRecord(mappingRecord(SERVER).settings).files);
  var type = key == "_file.m3u.id" ? "m3u" : "xmltv";
  var fileID = type == "xmltv" && value.lastIndexOf(".") > 0 ? value.substring(0, value.lastIndexOf(".")) : value;
  var provider = mappingRecord(mappingRecord(files[type])[fileID]);
  return mappingString(provider.name) || value;
}
function renderMappingCounts(rows) {
  var counts = document.createElement("p");
  counts.className = "tf-mapping-counts";
  counts.setAttribute("role", "status");
  counts.setAttribute("aria-live", "polite");
  counts.textContent = rows.length + " results \xB7 " + mappingWorkspaceState.selected.size + " selected (" + mappingSelectedVisibleCount(mappingWorkspaceState, rows.map(function(row) {
    return row.id;
  })) + " in results)";
  var selection = document.createElement("label");
  selection.className = "tf-mapping-select-results";
  var all = document.createElement("input");
  all.type = "checkbox";
  all.setAttribute("aria-label", "Select all visible channels");
  all.setAttribute("data-mapping-focus", "select-all-visible");
  var selected = mappingSelectedVisibleCount(mappingWorkspaceState, rows.map(function(row) {
    return row.id;
  }));
  all.checked = rows.length > 0 && selected == rows.length;
  all.indeterminate = selected > 0 && selected < rows.length;
  all.disabled = rows.length == 0;
  all.addEventListener("change", function() {
    mappingSelectVisible(mappingWorkspaceState, rows.map(function(row) {
      return row.id;
    }), all.checked);
    renderMappingWorkspaceIfOpen();
  });
  selection.appendChild(all);
  selection.appendChild(document.createTextNode("Select results"));
  counts.appendChild(selection);
  return counts;
}
function renderMappingBulkBar(rows) {
  var bar = document.createElement("div");
  bar.className = "tf-mapping-bulk-bar";
  var summary = document.createElement("strong");
  var visibleSelected = mappingSelectedVisibleIDs(mappingWorkspaceState, rows);
  summary.textContent = mappingWorkspaceState.selected.size + " selected (" + visibleSelected.length + " in results)";
  bar.appendChild(summary);
  var edit = mappingButton("Edit selected", function() {
    openMappingEditor(visibleSelected, bar);
  });
  edit.setAttribute("data-mapping-focus", "bulk-edit");
  edit.setAttribute("data-mapping-mutation", "true");
  edit.setAttribute("data-mapping-idle-disabled", String(visibleSelected.length == 0));
  edit.disabled = visibleSelected.length == 0 || mappingDraftIsLocked(mappingWorkspaceState);
  bar.appendChild(edit);
  var dummy = document.createElement("select");
  dummy.setAttribute("aria-label", "Dummy guide duration");
  dummy.setAttribute("data-mapping-focus", "bulk-dummy");
  mappingDummyPrograms.forEach(function(value) {
    var option = document.createElement("option");
    option.value = value;
    option.textContent = value.replace("_", " ");
    dummy.appendChild(option);
  });
  bar.appendChild(dummy);
  var applyDummy = mappingButton("Apply dummy guide", function() {
    mappingAssignDummy(mappingWorkspaceState, visibleSelected, dummy.value);
    renderMappingWorkspaceIfOpen();
  });
  applyDummy.setAttribute("data-mapping-focus", "bulk-apply-dummy");
  applyDummy.setAttribute("data-mapping-mutation", "true");
  applyDummy.setAttribute("data-mapping-idle-disabled", String(visibleSelected.length == 0));
  applyDummy.disabled = visibleSelected.length == 0 || mappingDraftIsLocked(mappingWorkspaceState);
  bar.appendChild(applyDummy);
  var clear = mappingButton("Clear selection", function() {
    mappingWorkspaceState.selected.clear();
    renderMappingWorkspaceIfOpen();
  });
  clear.setAttribute("data-mapping-focus", "bulk-clear");
  bar.appendChild(clear);
  return bar;
}
function renderMappingTable(rows) {
  if (rows.length == 0) {
    var empty = document.createElement("section");
    empty.className = "tf-mapping-empty";
    var title = document.createElement("h2");
    title.textContent = "No channels in this view";
    var copy = document.createElement("p");
    copy.textContent = "Change the search, filters, or Mapping view.";
    empty.appendChild(title);
    empty.appendChild(copy);
    return empty;
  }
  var scroll = document.createElement("div");
  scroll.className = "table-responsive tf-mapping-table-scroll";
  var table = document.createElement("table");
  table.className = "table lg:table-stacked tf-mapping-table";
  var caption = document.createElement("caption");
  caption.textContent = "Mapping channels";
  table.appendChild(caption);
  var head = document.createElement("thead"), header = document.createElement("tr");
  var selection = document.createElement("th");
  selection.scope = "col";
  selection.textContent = "Selection";
  header.appendChild(selection);
  [["number", "Channel"], ["name", "Name"], ["playlist", "Playlist"], ["group", "Group"], ["xmltv", "XMLTV / guide"], ["", "State"], ["", "Actions"]].forEach(function(entry) {
    var th = document.createElement("th");
    th.scope = "col";
    if (entry[0]) {
      var sort = entry[0];
      var button = mappingButton(entry[1], function() {
        if (mappingCurrentQuery.sort == sort)
          mappingCurrentQuery.descending = !mappingCurrentQuery.descending;
        else {
          mappingCurrentQuery.sort = sort;
          mappingCurrentQuery.descending = false;
        }
        renderMappingWorkspaceIfOpen();
      });
      button.setAttribute("data-mapping-focus", "sort-" + sort);
      button.setAttribute("tabindex", "-1");
      if ((mappingCurrentQuery.sort || "number") == sort)
        th.setAttribute("aria-sort", mappingCurrentQuery.descending ? "descending" : "ascending");
      th.appendChild(button);
    } else
      th.textContent = entry[1];
    header.appendChild(th);
  });
  head.appendChild(header);
  table.appendChild(head);
  var body = document.createElement("tbody");
  rows.forEach(function(row) {
    body.appendChild(renderMappingRow(row, rows));
  });
  table.appendChild(body);
  scroll.appendChild(table);
  return scroll;
}
function renderMappingRow(row, visibleRows) {
  var tr = document.createElement("tr");
  tr.setAttribute("data-mapping-id", row.id);
  var selected = document.createElement("td"), checkbox = document.createElement("input");
  checkbox.type = "checkbox";
  checkbox.checked = mappingWorkspaceState.selected.has(row.id);
  checkbox.setAttribute("aria-label", "Select " + (mappingString(row.channel["x-name"]) || row.id));
  checkbox.setAttribute("data-mapping-focus", "row-" + row.id);
  checkbox.addEventListener("click", function(event) {
    mappingSelectRange(mappingWorkspaceState, visibleRows.map(function(item) {
      return item.id;
    }), row.id, checkbox.checked, event.shiftKey);
    renderMappingWorkspaceIfOpen();
  });
  selected.appendChild(checkbox);
  tr.appendChild(selected);
  mappingAppendCell(tr, mappingString(row.channel["x-channelID"]));
  mappingAppendCell(tr, mappingString(row.channel["x-name"] || row.channel["tvg-name"] || row.id));
  mappingAppendCell(tr, mappingProviderLabel("_file.m3u.id", mappingString(row.channel["_file.m3u.id"])));
  mappingAppendCell(tr, mappingString(row.channel["x-group-title"] || row.channel["group-title"]));
  mappingAppendCell(tr, mappingProviderLabel("x-xmltv-file", mappingString(row.channel["x-xmltv-file"])) + " \xB7 " + mappingString(row.channel["x-mapping"]));
  var state = document.createElement("td");
  state.className = "tf-mapping-state";
  state.textContent = row.reasons.length ? row.reasons.map(mappingAttentionReasonLabel).join(" \xB7 ") : "Active";
  tr.appendChild(state);
  var action = document.createElement("td");
  var edit = mappingButton("Edit", function() {
    openMappingEditor([row.id], edit);
  });
  edit.setAttribute("data-mapping-edit", row.id);
  edit.setAttribute("data-mapping-focus", "edit-" + row.id);
  edit.setAttribute("data-mapping-mutation", "true");
  edit.setAttribute("data-mapping-idle-disabled", "false");
  edit.disabled = mappingDraftIsLocked(mappingWorkspaceState);
  action.appendChild(edit);
  tr.appendChild(action);
  var cellLabels = ["Selection", "Channel", "Name", "Playlist", "Group", "XMLTV / guide", "State", "Actions"];
  Array.prototype.forEach.call(tr.children, function(cell, index) {
    cell.setAttribute("data-cell", cellLabels[index]);
  });
  return tr;
}
function mappingAppendCell(row, value) {
  var cell = document.createElement("td");
  cell.textContent = value;
  row.appendChild(cell);
}
function renderMappingBackupOptions() {
  var datalist = document.createElement("datalist");
  datalist.id = "mapping-backup-options";
  var none = document.createElement("option");
  none.value = "-";
  datalist.appendChild(none);
  var names = /* @__PURE__ */ new Set();
  Object.keys(mappingWorkspaceState.draft).forEach(function(id) {
    var name = mappingString(mappingWorkspaceState.draft[id]["tvg-name"]);
    if (name)
      names.add(name);
  });
  Array.from(names).sort().forEach(function(name) {
    var option = document.createElement("option");
    option.value = name;
    datalist.appendChild(option);
  });
  return datalist;
}
function renderMappingSaveBar() {
  var bar = document.createElement("div");
  bar.className = "tf-mapping-save-bar";
  var dirty = mappingDirtyIDs(mappingWorkspaceState);
  bar.hidden = dirty.length == 0 && !mappingWorkspaceState.feedback;
  var status = document.createElement("p");
  status.setAttribute("role", "status");
  status.setAttribute("aria-live", "polite");
  status.textContent = mappingWorkspaceState.feedback || dirty.length + " staged channels";
  bar.appendChild(status);
  var save = mappingButton("Save mapping", function() {
    saveMappingDraft();
  });
  save.id = "mapping-save";
  save.setAttribute("data-mapping-focus", "save");
  save.disabled = dirty.length == 0 || mappingWorkspaceState.saveState != "idle";
  var discard = mappingButton("Discard draft", function() {
    if (mappingWorkspaceState.saveState == "idle") {
      mappingRevertDraft(mappingWorkspaceState);
      renderMappingWorkspaceIfOpen();
    }
  });
  discard.setAttribute("data-mapping-focus", "discard");
  discard.setAttribute("data-mapping-mutation", "true");
  discard.setAttribute("data-mapping-idle-disabled", String(dirty.length == 0));
  discard.disabled = dirty.length == 0 || mappingWorkspaceState.saveState != "idle";
  bar.appendChild(save);
  bar.appendChild(discard);
  return bar;
}
function mappingButton(label, listener) {
  var button = document.createElement("button");
  button.type = "button";
  button.textContent = label;
  button.addEventListener("click", listener);
  return button;
}
function updateMappingMutationControls() {
  if (!mappingWorkspaceState || typeof document.querySelectorAll != "function")
    return;
  var controls = document.querySelectorAll("[data-mapping-mutation]");
  var locked = mappingDraftIsLocked(mappingWorkspaceState);
  Array.prototype.forEach.call(controls, function(control) {
    control.disabled = locked || control.getAttribute("data-mapping-idle-disabled") == "true";
  });
}
function openMappingEditor(ids, invoker) {
  if (!ids.length || mappingDraftIsLocked(mappingWorkspaceState))
    return;
  closeMappingEditor();
  var first = mappingRecord(mappingWorkspaceState.draft[ids[0]]);
  var overlay = document.createElement("dialog");
  overlay.id = "mapping-editor";
  overlay.className = "2xl:drawer drawer-end tf-mapping-editor";
  overlay.setAttribute("role", "dialog");
  overlay.setAttribute("aria-modal", "true");
  overlay.setAttribute("aria-labelledby", "mapping-editor-title");
  var panel = document.createElement("div");
  panel.className = "drawer-body tf-mapping-editor-panel";
  var title = document.createElement("h2");
  title.id = "mapping-editor-title";
  title.textContent = ids.length == 1 ? "Edit " + (mappingString(first["x-name"]) || ids[0]) : "Edit " + ids.length + " channels";
  panel.appendChild(title);
  var fieldSearch = document.createElement("input");
  fieldSearch.type = "search";
  fieldSearch.setAttribute("aria-label", "Search editor fields");
  fieldSearch.placeholder = "Search fields";
  panel.appendChild(fieldSearch);
  var form = document.createElement("form");
  form.className = "tf-mapping-editor-form";
  form.addEventListener("submit", function(event) {
    event.preventDefault();
    applyMappingEditor(ids, form, invoker);
  });
  var basic = document.createElement("div");
  basic.className = "tf-mapping-editor-fields";
  basic.appendChild(mappingEditorCheckbox("Active", "x-active", first["x-active"] === true));
  basic.appendChild(mappingEditorInput(ids.length == 1 ? "Channel number" : "Starting channel number", ids.length == 1 ? "x-channelID" : "x-channels-start", ids.length == 1 ? mappingString(first["x-channelID"]) : ""));
  basic.appendChild(mappingEditorInput("Name", "x-name", mappingString(first["x-name"])));
  basic.appendChild(mappingEditorSelect("XMLTV source", "x-xmltv-file", ["-", "Threadfin Dummy"].concat(Object.keys(mappingWorkspaceState.xmltvMap).filter(function(file) {
    return file != "-" && file != "Threadfin Dummy";
  })), mappingString(first["x-xmltv-file"])));
  var mappingInput = mappingEditorEPGCombobox(mappingString(first["x-xmltv-file"]), mappingString(first["x-mapping"]));
  basic.appendChild(mappingInput);
  var mappingControl;
  var sourceControl = basic.querySelector('[name="x-xmltv-file"]');
  sourceControl.addEventListener("change", function() {
    sourceControl.setAttribute("data-changed", "true");
    var previous = mappingControl ? mappingControl.value : "-";
    if (mappingEPGCombobox)
      mappingEPGCombobox.dispose();
    var next = mappingEditorEPGCombobox(sourceControl.value, previous, false);
    basic.removeChild(mappingInput);
    basic.appendChild(next);
    mappingInput = next;
    initializeEPGCombobox(true);
    ppvField.hidden = mappingControl.value != "PPV";
  });
  function initializeEPGCombobox(changed = false) {
    var toggle = mappingInput.querySelector(".combobox-toggle");
    mappingEPGCombobox = new Combobox(toggle, { name: "x-mapping", search: true, searchNormalize: true });
    mappingControl = mappingInput.querySelector('[name="x-mapping"]');
    if (changed)
      mappingControl.setAttribute("data-changed", "true");
    toggle.addEventListener("change.bs.combobox", function() {
      mappingControl.setAttribute("data-changed", "true");
      ppvField.hidden = mappingControl.value != "PPV";
    });
  }
  form.appendChild(basic);
  var advanced = document.createElement("details");
  advanced.className = "tf-mapping-editor-advanced";
  var summary = document.createElement("summary");
  summary.textContent = "Advanced fields";
  advanced.appendChild(summary);
  var ppvField = mappingEditorInput("PPV extra", "x-ppv-extra", mappingString(first["x-ppv-extra"]));
  ppvField.hidden = mappingString(first["x-mapping"]) != "PPV";
  [
    mappingEditorInput("Description", "x-description", mappingString(first["x-description"])),
    mappingEditorCheckbox("Update channel name", "x-update-channel-name", first["x-update-channel-name"] === true),
    mappingEditorInput("Logo", "tvg-logo", mappingString(first["tvg-logo"])),
    mappingEditorCheckbox("Update channel logo", "x-update-channel-icon", first["x-update-channel-icon"] === true),
    mappingEditorInput("Category", "x-category", mappingString(first["x-category"])),
    mappingEditorInput("Group", "x-group-title", mappingString(first["x-group-title"])),
    ppvField,
    mappingEditorInput("Backup channel 1", "x-backup-channel-1", mappingString(first["x-backup-channel-1"]), "mapping-backup-options"),
    mappingEditorInput("Backup channel 2", "x-backup-channel-2", mappingString(first["x-backup-channel-2"]), "mapping-backup-options"),
    mappingEditorInput("Backup channel 3", "x-backup-channel-3", mappingString(first["x-backup-channel-3"]), "mapping-backup-options"),
    mappingEditorCheckbox("Hidden from outputs", "x-hide-channel", first["x-hide-channel"] === true)
  ].forEach(function(field) {
    advanced.appendChild(field);
  });
  form.appendChild(advanced);
  var probe = mappingButton("Probe channel", function() {
    requestMappingProbe(mappingString(first.url));
  });
  probe.type = "button";
  form.appendChild(probe);
  var probeStatus = document.createElement("p");
  probeStatus.id = "mapping-probe-status";
  probeStatus.setAttribute("role", "status");
  probeStatus.setAttribute("aria-live", "polite");
  probeStatus.textContent = mappingProbeFeedback;
  form.appendChild(probeStatus);
  var formStatus = document.createElement("p");
  formStatus.id = "mapping-editor-status";
  formStatus.setAttribute("role", "alert");
  form.appendChild(formStatus);
  var actions = document.createElement("div");
  actions.className = "tf-mapping-editor-actions";
  actions.appendChild(mappingButton("Cancel", function() {
    closeMappingEditor(invoker);
  }));
  var apply = mappingButton("Apply to draft", function() {
    applyMappingEditor(ids, form, invoker);
  });
  apply.id = "mapping-apply";
  actions.appendChild(apply);
  form.appendChild(actions);
  apply.setAttribute("data-mapping-mutation", "true");
  apply.setAttribute("data-mapping-idle-disabled", "false");
  apply.disabled = mappingDraftIsLocked(mappingWorkspaceState);
  fieldSearch.addEventListener("input", function() {
    filterMappingEditorFields(form, advanced, fieldSearch.value);
  });
  panel.appendChild(form);
  overlay.appendChild(panel);
  mappingOverlayHost().appendChild(overlay);
  initializeEPGCombobox();
  overlay.addEventListener("hidden.bs.drawer", function() {
    if (window.matchMedia("(min-width: 1536px)").matches)
      return;
    closeMappingEditor(invoker);
  });
  mappingEditorDrawer = new Drawer(overlay);
  mappingEditorBreakpoint = window.matchMedia("(min-width: 1536px)");
  mappingEditorResize = function(event) {
    overlay.setAttribute("aria-modal", String(!event.matches));
    if (!event.matches && mappingEditorDrawer)
      mappingEditorDrawer.show();
  };
  mappingEditorBreakpoint.addEventListener("change", mappingEditorResize);
  var firstField = form.querySelector("input, select");
  if (window.matchMedia("(min-width: 1536px)").matches) {
    overlay.setAttribute("aria-modal", "false");
    if (firstField)
      firstField.focus();
  } else {
    mappingEditorDrawer.show().then(function() {
      if (firstField)
        firstField.focus();
    });
  }
}
function mappingEditorInput(label, name, value, list) {
  var wrapper = document.createElement("label");
  wrapper.className = "form-field tf-mapping-editor-field";
  wrapper.setAttribute("data-label", label.toLocaleLowerCase());
  wrapper.textContent = label;
  var input = document.createElement("input");
  input.type = "text";
  input.name = name;
  input.value = value;
  if (list)
    input.setAttribute("list", list);
  input.addEventListener("input", function() {
    input.setAttribute("data-changed", "true");
  });
  wrapper.appendChild(input);
  return wrapper;
}
function mappingEditorCheckbox(label, name, checked) {
  var wrapper = document.createElement("label");
  wrapper.className = "form-field tf-mapping-editor-field tf-mapping-editor-checkbox";
  wrapper.setAttribute("data-label", label.toLocaleLowerCase());
  var input = document.createElement("input");
  input.type = "checkbox";
  input.name = name;
  input.checked = checked;
  input.addEventListener("change", function() {
    input.setAttribute("data-changed", "true");
  });
  wrapper.appendChild(input);
  wrapper.appendChild(document.createTextNode(label));
  return wrapper;
}
function mappingEditorSelect(label, name, values, selected) {
  var wrapper = document.createElement("label");
  wrapper.className = "form-field tf-mapping-editor-field";
  wrapper.setAttribute("data-label", label.toLocaleLowerCase());
  wrapper.textContent = label;
  var select = document.createElement("select");
  select.name = name;
  var choices = values.indexOf(selected) == -1 && selected ? values.concat([selected]) : values;
  choices.forEach(function(value) {
    var option = document.createElement("option");
    option.value = value;
    option.textContent = mappingProviderLabel("x-xmltv-file", value) + (values.indexOf(value) == -1 ? " (unavailable)" : "");
    select.appendChild(option);
  });
  select.value = selected;
  select.addEventListener("change", function() {
    select.setAttribute("data-changed", "true");
  });
  wrapper.appendChild(select);
  return wrapper;
}
function filterMappingEditorFields(form, advanced, value) {
  var search = value.trim().toLocaleLowerCase();
  var fields = form.querySelectorAll(".tf-mapping-editor-field");
  var advancedMatch = false;
  Array.prototype.forEach.call(fields, function(field) {
    var match = !search || mappingString(field.getAttribute("data-label")).indexOf(search) != -1;
    field.hidden = !match;
    if (match && advanced.contains(field))
      advancedMatch = true;
  });
  if (search && advancedMatch)
    advanced.open = true;
}
function applyMappingEditor(ids, form, invoker) {
  var patch = {};
  var fields = form.querySelectorAll("input[name], select[name]");
  Array.prototype.forEach.call(fields, function(field) {
    if (ids.length > 1 && field.getAttribute("data-changed") != "true")
      return;
    if (field.name == "x-channels-start")
      return;
    patch[field.name] = field instanceof HTMLInputElement && field.type == "checkbox" ? field.checked : field.value;
  });
  var number = form.querySelector('[name="x-channelID"]');
  var start = form.querySelector('[name="x-channels-start"]');
  var result = mappingApplyChannelPatch(mappingWorkspaceState, ids, patch, { numberChanged: !!number && number.getAttribute("data-changed") == "true", sequentialStart: start && start.getAttribute("data-changed") == "true" ? start.value : void 0 });
  if (!result.ok) {
    var status = document.getElementById("mapping-editor-status");
    if (status)
      status.textContent = result.error;
    if (number)
      number.focus();
    else if (start)
      start.focus();
    return;
  }
  closeMappingEditor();
  renderMappingWorkspaceIfOpen();
  var save = document.getElementById("mapping-save");
  if (save)
    save.focus();
}
function closeMappingEditor(invoker) {
  if (mappingEditorBreakpoint && mappingEditorResize)
    mappingEditorBreakpoint.removeEventListener("change", mappingEditorResize);
  mappingEditorBreakpoint = void 0;
  mappingEditorResize = void 0;
  if (mappingEPGCombobox) {
    mappingEPGCombobox.dispose();
    mappingEPGCombobox = void 0;
  }
  if (mappingEditorDrawer) {
    mappingEditorDrawer.dispose();
    mappingEditorDrawer = void 0;
  }
  var editor = document.getElementById("mapping-editor");
  if (editor && editor.parentElement)
    editor.parentElement.removeChild(editor);
  if (invoker && document.contains(invoker))
    invoker.focus();
}
function mappingEditorEPGCombobox(source, current, preserveUnavailable = true) {
  var wrapper = document.createElement("div");
  wrapper.className = "form-field tf-mapping-editor-field";
  wrapper.setAttribute("data-label", "xmltv channel");
  var label = document.createElement("label");
  label.id = "mapping-epg-label";
  label.textContent = "XMLTV channel";
  var toggle = document.createElement("button");
  toggle.type = "button";
  toggle.className = "form-control combobox-toggle";
  toggle.setAttribute("data-bs-toggle", "combobox");
  toggle.setAttribute("aria-labelledby", label.id);
  var value = document.createElement("span");
  value.className = "combobox-value";
  toggle.appendChild(value);
  var menu = document.createElement("div");
  menu.className = "menu";
  var search = document.createElement("input");
  search.type = "search";
  search.className = "combobox-search-input";
  search.placeholder = "Search channel name or ID";
  search.setAttribute("aria-label", "Search EPG channels");
  menu.appendChild(search);
  var programs = mappingRecord(mappingWorkspaceState.xmltvMap[source]);
  var choices = ["-"].concat(source == "Threadfin Dummy" ? mappingDummyPrograms : Object.keys(programs));
  var unavailable = !!current && choices.indexOf(current) == -1;
  if (preserveUnavailable && unavailable)
    choices.push(current);
  var selected = choices.indexOf(current) != -1 ? current : "-";
  choices.forEach(function(id) {
    var option = document.createElement("button");
    option.type = "button";
    option.className = "menu-item" + (selected == id ? " selected" : "");
    option.setAttribute("data-bs-value", id);
    option.setAttribute("aria-selected", String(selected == id));
    var name = mappingString(mappingRecord(programs[id])["display-name"]);
    option.textContent = (name && name != id ? name + " \xB7 " + id : id) + (unavailable && id == current ? " (unavailable)" : "");
    menu.appendChild(option);
  });
  var noResults = document.createElement("p");
  noResults.className = "combobox-no-results d-none";
  noResults.textContent = "No matching EPG channels";
  menu.appendChild(noResults);
  wrapper.appendChild(label);
  wrapper.appendChild(toggle);
  wrapper.appendChild(menu);
  return wrapper;
}
var sourceFeedbackByKey = {};
var sourcePageFeedback = {};
var sourcePopupInvoker;
var sourcePopupFocusKey = "";
var sourcePopupFocusListenerAttached = false;
var sourcePopupShownListenerAttached = false;
function sourceRecord(value) {
  return value && typeof value == "object" && !Array.isArray(value) ? value : {};
}
function sourceString(value) {
  return typeof value == "string" ? value : value === void 0 || value === null ? "" : String(value);
}
function sourceNumber(value) {
  var number = typeof value == "number" ? value : Number(value);
  return isFinite(number) ? number : 0;
}
function sourceLocationForDisplay(value) {
  var source = sourceString(value);
  if (!/^https?:/i.test(source.trim())) {
    return source;
  }
  try {
    var location2 = new URL(source.trim());
    if (!location2.protocol || !location2.host) {
      return "{{.sources.locationInvalidDisplay}}";
    }
    return location2.protocol + "//" + location2.host + location2.pathname;
  } catch (_error) {
    return "{{.sources.locationInvalidDisplay}}";
  }
}
function sourceFeedbackKey(providerType, id) {
  return providerType + ":" + id;
}
function selectSourceList(server, destination) {
  var root = sourceRecord(server);
  var settings = sourceRecord(root.settings);
  var files = sourceRecord(settings.files);
  var types = destination == "playlist" ? ["m3u", "hdhr"] : ["xmltv"];
  var sources = [];
  types.forEach((providerType) => {
    var configured = sourceRecord(files[providerType]);
    Object.keys(configured).forEach((id) => {
      var source = sourceRecord(configured[id]);
      var compatibility = sourceRecord(source.compatibility);
      var rawAvailability = source["provider.availability"];
      var counts = [];
      if (providerType == "xmltv") {
        counts.push({ label: "{{.sources.counts.channels}}", value: sourceString(compatibility["xmltv.channels"] || 0) });
        counts.push({ label: "{{.sources.counts.programs}}", value: sourceString(compatibility["xmltv.programs"] || 0) });
      } else {
        counts.push({ label: "{{.sources.counts.streams}}", value: sourceString(compatibility.streams || 0) });
        counts.push({ label: "{{.sources.counts.tuners}}", value: sourceString(source.tuner || 0) });
        counts.push({ label: "{{.sources.counts.groupCoverage}}", value: sourceString(compatibility["group.title"] || 0) + "%" });
        counts.push({ label: "{{.sources.counts.tvgIDCoverage}}", value: sourceString(compatibility["tvg.id"] || 0) + "%" });
        counts.push({ label: "{{.sources.counts.uniqueIDCoverage}}", value: sourceString(compatibility["stream.id"] || 0) + "%" });
      }
      sources.push({
        id,
        providerType,
        name: sourceString(source.name) || (providerType == "xmltv" ? "XMLTV" : providerType.toUpperCase()),
        typeLabel: providerType == "m3u" ? "M3U" : providerType == "hdhr" ? "HDHomeRun" : "XMLTV",
        source: sourceLocationForDisplay(source["file.source"]),
        lastUpdate: sourceString(source["last.update"]),
        availability: sourceNumber(rawAvailability),
        availabilityKnown: rawAvailability !== void 0 && rawAvailability !== null,
        counts
      });
    });
  });
  sources.sort((a, b) => {
    var byName = a.name.localeCompare(b.name);
    return byName == 0 ? a.id.localeCompare(b.id) : byName;
  });
  return sources;
}
function renderSourceManagementPage(destination, host) {
  host.innerHTML = "";
  var page = document.createElement("div");
  page.className = "tf-sources";
  page.setAttribute("data-source-destination", destination);
  var header = document.createElement("header");
  header.className = "tf-sources-header";
  var titleGroup = document.createElement("div");
  var title = document.createElement("h1");
  title.textContent = destination == "playlist" ? "{{.mainMenu.item.playlist}}" : "{{.mainMenu.item.xmltv}}";
  var purpose = document.createElement("p");
  purpose.textContent = destination == "playlist" ? "{{.sources.playlist.purpose}}" : "{{.sources.xmltv.purpose}}";
  titleGroup.appendChild(title);
  titleGroup.appendChild(purpose);
  header.appendChild(titleGroup);
  header.appendChild(createSourceAddButton(destination));
  page.appendChild(header);
  if (destination == "xmltv" && sourceString(sourceRecord(SERVER["clientInfo"]).epgSource) != "XEPG") {
    var modeNote = document.createElement("div");
    modeNote.className = "tf-source-mode-note";
    modeNote.textContent = "{{.sources.xmltv.pmsNote}}";
    var settingsAction = document.createElement("button");
    settingsAction.type = "button";
    settingsAction.textContent = "{{.sources.xmltv.reviewEpgSource}}";
    settingsAction.addEventListener("click", function() {
      openDestination("settings", true);
    });
    modeNote.appendChild(settingsAction);
    page.appendChild(modeNote);
  }
  page.appendChild(renderSourcePageFeedback(destination));
  var sources = selectSourceList(SERVER, destination);
  if (sources.length == 0) {
    page.appendChild(renderSourceEmptyState(destination));
  } else {
    var list = document.createElement("div");
    list.className = "tf-source-list";
    list.setAttribute("role", "list");
    sources.forEach((source) => list.appendChild(renderSourceRow(source)));
    page.appendChild(list);
  }
  host.appendChild(page);
}
function createSourceAddButton(destination) {
  var button = document.createElement("button");
  button.type = "button";
  button.className = "tf-source-primary-action";
  button.setAttribute("data-source-focus-key", destination + ":add");
  button.textContent = destination == "playlist" ? "{{.sources.playlist.add}}" : "{{.sources.xmltv.add}}";
  button.addEventListener("click", function() {
    openSourcePopup(destination == "playlist" ? "playlist" : "xmltv", void 0, button);
  });
  return button;
}
function renderSourcePageFeedback(destination) {
  var region = document.createElement("div");
  region.id = "source-page-status";
  region.className = "tf-source-page-status";
  region.setAttribute("role", "status");
  region.setAttribute("aria-live", "polite");
  var feedback = sourcePageFeedback[destination];
  if (!feedback) {
    region.hidden = true;
    return region;
  }
  region.setAttribute("data-state", feedback.state);
  var message = document.createElement("span");
  message.textContent = feedback.message;
  region.appendChild(message);
  if (feedback.nextDestination && feedback.nextLabel) {
    var next = document.createElement("button");
    next.type = "button";
    next.textContent = feedback.nextLabel;
    next.addEventListener("click", function() {
      openDestination(feedback.nextDestination, true);
    });
    region.appendChild(next);
  }
  return region;
}
function renderSourceEmptyState(destination) {
  var empty = document.createElement("section");
  empty.className = "tf-source-empty";
  var title = document.createElement("h2");
  title.textContent = destination == "playlist" ? "{{.sources.playlist.emptyTitle}}" : "{{.sources.xmltv.emptyTitle}}";
  var explanation = document.createElement("p");
  explanation.textContent = destination == "playlist" ? "{{.sources.playlist.emptyDescription}}" : "{{.sources.xmltv.emptyDescription}}";
  empty.appendChild(title);
  empty.appendChild(explanation);
  empty.appendChild(createSourceAddButton(destination));
  return empty;
}
function renderSourceRow(source) {
  var row = document.createElement("article");
  row.className = "tf-source-row";
  row.setAttribute("role", "listitem");
  row.setAttribute("data-source-id", source.id);
  row.setAttribute("data-source-type", source.providerType);
  var identity = document.createElement("div");
  identity.className = "tf-source-identity";
  var name = document.createElement("h2");
  name.textContent = source.name;
  var type = document.createElement("p");
  type.className = "tf-source-type";
  type.textContent = source.typeLabel;
  var location2 = document.createElement("code");
  location2.textContent = source.source || "{{.sources.locationUnavailable}}";
  identity.appendChild(name);
  identity.appendChild(type);
  identity.appendChild(location2);
  row.appendChild(identity);
  var health = document.createElement("div");
  health.className = "tf-source-health";
  var availability = document.createElement("strong");
  availability.className = "tf-source-availability";
  if (!source.availabilityKnown) {
    availability.textContent = "{{.sources.availabilityUnknown}}";
    health.setAttribute("data-status", "unknown");
  } else if (source.availability > 0) {
    availability.textContent = "{{.sources.available}} \xB7 " + source.availability + "%";
    health.setAttribute("data-status", "ready");
  } else {
    availability.textContent = "{{.sources.unavailable}} \xB7 " + source.availability + "%";
    health.setAttribute("data-status", "unavailable");
  }
  var updated = document.createElement("span");
  updated.textContent = source.lastUpdate ? "{{.sources.lastUpdate}}: " + source.lastUpdate : "{{.sources.neverUpdated}}";
  health.appendChild(availability);
  health.appendChild(updated);
  row.appendChild(health);
  var counts = document.createElement("dl");
  counts.className = "tf-source-counts";
  source.counts.forEach((count) => {
    var group = document.createElement("div");
    var term = document.createElement("dt");
    term.textContent = count.label;
    var value = document.createElement("dd");
    value.textContent = count.value;
    group.appendChild(term);
    group.appendChild(value);
    counts.appendChild(group);
  });
  row.appendChild(counts);
  var feedback = sourceFeedbackByKey[sourceFeedbackKey(source.providerType, source.id)];
  var feedbackElement = document.createElement("p");
  feedbackElement.className = "tf-source-feedback";
  feedbackElement.setAttribute("role", "status");
  feedbackElement.setAttribute("aria-live", "polite");
  if (feedback) {
    feedbackElement.textContent = feedback.message;
    feedbackElement.setAttribute("data-state", feedback.state);
  } else {
    feedbackElement.hidden = true;
  }
  row.appendChild(feedbackElement);
  var actions = document.createElement("div");
  actions.className = "tf-source-actions";
  actions.appendChild(createSourceRowAction("{{.sources.actions.edit}}", source.providerType + ":" + source.id + ":edit", function(invoker) {
    openSourcePopup(source.providerType, { id: source.id }, invoker);
  }));
  actions.appendChild(createSourceRowAction("{{.sources.actions.update}}", source.providerType + ":" + source.id + ":update", function(invoker) {
    invokeSourcePopupAction(source.providerType, source.id, "update", invoker);
  }));
  var remove = createSourceRowAction("{{.sources.actions.delete}}", source.providerType + ":" + source.id + ":delete", function(invoker) {
    invokeSourcePopupAction(source.providerType, source.id, "delete", invoker);
  });
  remove.className += " tf-source-delete-action";
  actions.appendChild(remove);
  row.appendChild(actions);
  return row;
}
function createSourceRowAction(label, focusKey, listener) {
  var button = document.createElement("button");
  button.type = "button";
  button.textContent = label;
  button.setAttribute("data-source-focus-key", focusKey);
  button.addEventListener("click", function() {
    listener(button);
  });
  return button;
}
function invokeSourcePopupAction(providerType, id, action, invoker) {
  openSourcePopup(providerType, { id }, invoker);
  savePopupData(providerType, id, action == "delete", action == "update" ? 1 : 0);
}
function openSourcePopup(dataType, element, invoker) {
  sourcePopupInvoker = invoker;
  sourcePopupFocusKey = invoker ? sourceString(invoker.getAttribute("data-source-focus-key")) : "";
  var modal = document.getElementById("popup");
  if (modal) {
    modal.setAttribute("role", "dialog");
    modal.setAttribute("aria-modal", "true");
  }
  if (modal && !sourcePopupFocusListenerAttached) {
    sourcePopupFocusListenerAttached = true;
    modal.addEventListener("hidden.bs.dialog", function() {
      var focusTarget = sourcePopupInvoker && document.contains(sourcePopupInvoker) ? sourcePopupInvoker : sourcePopupReplacement(sourcePopupFocusKey);
      if (focusTarget) {
        focusTarget.focus();
      }
      sourcePopupInvoker = void 0;
      sourcePopupFocusKey = "";
    });
  }
  if (modal && !sourcePopupShownListenerAttached) {
    sourcePopupShownListenerAttached = true;
    modal.addEventListener("shown.bs.dialog", function() {
      focusSourcePopupFirstControl();
    });
  }
  openPopUp(dataType, element);
}
function focusSourcePopupFirstControl() {
  var popup = document.getElementById("popup-custom");
  if (!popup || !popup.classList.contains("tf-source-popup")) {
    return;
  }
  var firstControl = popup.querySelector("input, select, button");
  if (firstControl) {
    firstControl.focus();
  }
}
function focusRenderedSourcePopup(modal) {
  if (modal && modal.open) {
    window.setTimeout(focusSourcePopupFirstControl, 0);
  }
}
function sourcePopupReplacement(focusKey) {
  if (!focusKey) {
    return void 0;
  }
  var candidates = document.querySelectorAll("[data-source-focus-key]");
  for (var index = 0; index < candidates.length; index++) {
    var candidate = candidates[index];
    if (candidate.getAttribute("data-source-focus-key") == focusKey) {
      return candidate;
    }
  }
  return void 0;
}
function enhanceSourcePopup(dataType) {
  var popup = document.getElementById("popup-custom");
  if (!popup) {
    return;
  }
  popup.classList.remove("tf-source-popup");
  var sourcePopup = dataType == "playlist" || dataType == "m3u" || dataType == "hdhr" || dataType == "xmltv";
  var title = popup.querySelector("h3");
  var modal = document.getElementById("popup");
  if (modal) {
    if (title) {
      title.id = sourcePopup ? "source-popup-title" : "popup-title";
      modal.setAttribute("aria-labelledby", title.id);
    } else {
      modal.removeAttribute("aria-labelledby");
    }
  }
  if (!sourcePopup) {
    return;
  }
  popup.classList.add("tf-source-popup");
  if (dataType == "playlist") {
    focusRenderedSourcePopup(modal);
    return;
  }
  var fields = popup.querySelectorAll("input, select");
  Array.prototype.forEach.call(fields, function(field) {
    var row = field.closest("tr");
    var title2 = row ? row.querySelector("td:first-child") : void 0;
    if (title2 && title2.textContent) {
      field.setAttribute("aria-label", title2.textContent.replace(/:\s*$/, ""));
    }
  });
  var sourceInput = popup.querySelector('[name="file.source"]');
  if (sourceInput && sourceInput.parentElement) {
    var help = document.createElement("p");
    help.id = "source-location-help";
    help.className = "tf-source-field-help";
    help.textContent = dataType == "hdhr" ? "{{.sources.forms.hdhrHelp}}" : "{{.sources.forms.locationHelp}}";
    sourceInput.setAttribute("aria-describedby", help.id + " source-file-source-error");
    sourceInput.parentElement.appendChild(help);
  }
  ;
  ["name", "file.source"].forEach((fieldName) => {
    var field = popup.querySelector('[name="' + fieldName + '"]');
    if (!field || !field.parentElement) {
      return;
    }
    field.setAttribute("aria-required", "true");
    var error = document.createElement("p");
    error.id = "source-" + fieldName.replace(".", "-") + "-error";
    error.className = "tf-source-field-error";
    error.hidden = true;
    field.parentElement.appendChild(error);
  });
  var status = document.createElement("p");
  status.id = "source-form-status";
  status.className = "tf-source-form-status";
  status.setAttribute("role", "status");
  status.setAttribute("aria-live", "polite");
  status.hidden = true;
  popup.appendChild(status);
  focusRenderedSourcePopup(modal);
}
function sourceLocationAccepted(value) {
  var source = sourceString(value).trim();
  if (!source) {
    return false;
  }
  var scheme = source.match(/^([a-z][a-z0-9+.-]*):\/\//i);
  return !scheme || scheme[1].toLowerCase() == "http" || scheme[1].toLowerCase() == "https";
}
function validateSourcePopup(dataType) {
  if (dataType != "m3u" && dataType != "hdhr" && dataType != "xmltv") {
    return true;
  }
  var popup = document.getElementById("popup-custom");
  if (!popup) {
    return false;
  }
  sourceClearFormErrors();
  var name = popup.querySelector('[name="name"]');
  var source = popup.querySelector('[name="file.source"]');
  if (!name || !name.value.trim()) {
    sourceSetFieldError(name, "source-name-error", "{{.sources.forms.nameRequired}}");
    return false;
  }
  if (!source || !source.value.trim()) {
    sourceSetFieldError(source, "source-file-source-error", "{{.sources.forms.locationRequired}}");
    return false;
  }
  if (dataType != "hdhr" && !sourceLocationAccepted(source.value)) {
    sourceSetFieldError(source, "source-file-source-error", "{{.sources.forms.locationInvalid}}");
    return false;
  }
  return true;
}
function sourceClearFormErrors() {
  var popup = document.getElementById("popup-custom");
  if (!popup) {
    return;
  }
  var invalid = popup.querySelectorAll('[aria-invalid="true"]');
  Array.prototype.forEach.call(invalid, function(field) {
    field.removeAttribute("aria-invalid");
  });
  var errors = popup.querySelectorAll(".tf-source-field-error");
  Array.prototype.forEach.call(errors, function(error) {
    error.textContent = "";
    error.hidden = true;
  });
}
function sourceSetFieldError(field, errorID, message) {
  if (field) {
    field.setAttribute("aria-invalid", "true");
  }
  var error = document.getElementById(errorID);
  if (error) {
    error.textContent = message;
    error.hidden = false;
  }
  if (field) {
    field.focus();
  }
}
function sourceSetFormStatus(message, state) {
  var status = document.getElementById("source-form-status");
  if (!status) {
    return;
  }
  status.textContent = message;
  status.setAttribute("data-state", state);
  status.hidden = false;
}
function beginSourceRequest(dataType, id, remove, option) {
  if (dataType != "m3u" && dataType != "hdhr" && dataType != "xmltv") {
    return;
  }
  var providerType = dataType;
  var message = remove == true ? "Deleting source\u2026" : option == 1 ? "Updating source\u2026" : id == "-" ? "Adding source\u2026" : "Saving source\u2026";
  var feedback = { state: "progress", message };
  sourceFeedbackByKey[sourceFeedbackKey(providerType, id)] = feedback;
  sourceSetFormStatus(message, "progress");
}
function sourceRequestDescriptor(command, data) {
  var commandMap = {
    saveFilesM3U: { providerType: "m3u", mode: "save" },
    updateFileM3U: { providerType: "m3u", mode: "update" },
    saveFilesHDHR: { providerType: "hdhr", mode: "save" },
    updateFileHDHR: { providerType: "hdhr", mode: "update" },
    saveFilesXMLTV: { providerType: "xmltv", mode: "save" },
    updateFileXMLTV: { providerType: "xmltv", mode: "update" }
  };
  var contract = commandMap[command];
  if (!contract) {
    return void 0;
  }
  var files = sourceRecord(sourceRecord(data).files);
  var providers = sourceRecord(files[contract.providerType]);
  var ids = Object.keys(providers);
  if (ids.length != 1) {
    return void 0;
  }
  var values = sourceRecord(providers[ids[0]]);
  return {
    providerType: contract.providerType,
    destination: contract.providerType == "xmltv" ? "xmltv" : "playlist",
    id: ids[0],
    mode: values.delete === true ? "delete" : contract.mode,
    values
  };
}
function completeSourceRequest(command, data, response) {
  var request = sourceRequestDescriptor(command, data);
  if (!request) {
    return;
  }
  var root = sourceRecord(response);
  if (root.status !== true) {
    var error = root.status === false ? sourceString(root.err) || "Source request failed." : "{{.sources.responseInvalid}}";
    var failed = { state: "error", message: error };
    sourceFeedbackByKey[sourceFeedbackKey(request.providerType, request.id)] = failed;
    sourceSetFormStatus(error, "error");
    return;
  }
  if (request.mode == "delete") {
    delete sourceFeedbackByKey[sourceFeedbackKey(request.providerType, request.id)];
    sourcePageFeedback[request.destination] = { state: "success", message: "Source deleted." };
    return;
  }
  var responseFiles = sourceRecord(sourceRecord(sourceRecord(root.settings).files)[request.providerType]);
  var resolvedID = request.id;
  if (resolvedID == "-") {
    Object.keys(responseFiles).some((id) => {
      var candidate = sourceRecord(responseFiles[id]);
      if (sourceString(candidate["file.source"]) == sourceString(request.values["file.source"]) && sourceString(candidate.name) == sourceString(request.values.name)) {
        resolvedID = id;
        return true;
      }
      return false;
    });
  }
  var saved = sourceRecord(responseFiles[resolvedID]);
  var lastUpdate = sourceString(saved["last.update"]);
  var message = request.mode == "update" ? "Update completed" + (lastUpdate ? ". Last update: " + lastUpdate + "." : ".") : "Source saved.";
  sourceFeedbackByKey[sourceFeedbackKey(request.providerType, resolvedID)] = { state: "success", message };
  if (request.id == "-") {
    sourcePageFeedback[request.destination] = request.destination == "playlist" ? {
      state: "success",
      message: "Playlist added. Next, select imported streams with a Filter.",
      nextDestination: "filter",
      nextLabel: "Open Filter"
    } : {
      state: "success",
      message: "XMLTV source added. Next, review guide assignments in Mapping.",
      nextDestination: "mapping",
      nextLabel: "Open Mapping"
    };
  }
}
function overviewRecord(value) {
  if (value && typeof value == "object" && !Array.isArray(value)) {
    return value;
  }
  return {};
}
function overviewNumber(value) {
  var parsed = typeof value == "number" ? value : Number(value);
  return isFinite(parsed) && parsed > 0 ? Math.floor(parsed) : 0;
}
function overviewString(value) {
  return typeof value == "string" ? value.trim() : "";
}
function overviewObjectCount(value) {
  return Object.keys(overviewRecord(value)).length;
}
function selectPlaylistCount(server) {
  var settings = overviewRecord(overviewRecord(server).settings);
  var files = overviewRecord(settings.files);
  return overviewObjectCount(files.m3u) + overviewObjectCount(files.hdhr);
}
function selectSelectedStreamCount(server) {
  var clientInfo = overviewRecord(overviewRecord(server).clientInfo);
  var streamCount = overviewString(clientInfo.streams).match(/^\s*(\d+)/);
  if (streamCount) {
    return overviewNumber(streamCount[1]);
  }
  var data = overviewRecord(overviewRecord(server).data);
  var preview = overviewRecord(data.StreamPreviewUI);
  return Array.isArray(preview.activeStreams) ? preview.activeStreams.length : 0;
}
function overviewXMLTVSourceReady(source) {
  var compatibility = overviewRecord(source.compatibility);
  return overviewNumber(source["provider.availability"]) > 0 && overviewNumber(compatibility["xmltv.channels"]) > 0;
}
function selectXMLTVState(server) {
  var root = overviewRecord(server);
  var settings = overviewRecord(root.settings);
  var clientInfo = overviewRecord(root.clientInfo);
  var files = overviewRecord(settings.files);
  var sources = overviewRecord(files.xmltv);
  var applicable = overviewString(settings.epgSource || clientInfo.epgSource).toUpperCase() == "XEPG";
  var sourceCount = Object.keys(sources).length;
  var readyCount = 0;
  Object.keys(sources).forEach((key) => {
    var source = overviewRecord(sources[key]);
    if (overviewXMLTVSourceReady(source)) {
      readyCount++;
    }
  });
  return {
    applicable,
    ready: applicable && sourceCount > 0 && readyCount == sourceCount,
    sourceCount,
    readyCount
  };
}
function selectMappingState(server) {
  var root = overviewRecord(server);
  var xepg = overviewRecord(root.xepg);
  var mappings = overviewRecord(xepg.epgMapping);
  var activeCount = 0;
  var unresolvedCount = 0;
  Object.keys(mappings).forEach((key) => {
    var mapping = overviewRecord(mappings[key]);
    if (mapping["x-hide-channel"] === true) {
      return;
    }
    var xmltvFile = overviewString(mapping["x-xmltv-file"]);
    var mappedChannel = overviewString(mapping["x-mapping"]);
    var valid = xmltvFile != "" && xmltvFile != "-" && mappedChannel != "" && mappedChannel != "-";
    if (mapping["x-active"] === true && valid) {
      activeCount++;
    } else if (!valid) {
      unresolvedCount++;
    }
  });
  return { activeCount, unresolvedCount };
}
function overviewEndpointAvailable(value, endpoint) {
  var address = overviewString(value);
  if (!address) {
    return false;
  }
  if (endpoint == "dvr") {
    return !/\s/.test(address);
  }
  return /^https?:\/\//i.test(address);
}
function selectOutputState(server, selectedStreamCount, xmltv, mapping) {
  var root = overviewRecord(server);
  var clientInfo = overviewRecord(root.clientInfo);
  var dvrAvailable = selectedStreamCount > 0 && overviewEndpointAvailable(clientInfo.DVR, "dvr");
  var m3uAvailable = xmltv.applicable && xmltv.ready && mapping.activeCount > 0 && overviewEndpointAvailable(clientInfo["m3u-url"], "m3u");
  var xmltvAvailable = xmltv.applicable && xmltv.ready && mapping.activeCount > 0 && overviewEndpointAvailable(clientInfo["xepg-url"], "xmltv");
  var ready = dvrAvailable;
  if (xmltv.applicable) {
    ready = ready && xmltv.ready && mapping.activeCount > 0 && m3uAvailable && xmltvAvailable;
  }
  return {
    ready,
    endpoints: [
      {
        key: "dvr",
        label: "DVR address",
        value: dvrAvailable ? overviewString(clientInfo.DVR) : "",
        available: dvrAvailable,
        explanation: dvrAvailable ? "Use this address when a client discovers Threadfin as a tuner." : selectedStreamCount == 0 ? "The DVR address becomes usable after streams are selected." : "The DVR address is not available yet."
      },
      {
        key: "m3u",
        label: "M3U URL",
        value: m3uAvailable ? overviewString(clientInfo["m3u-url"]) : "",
        available: m3uAvailable,
        explanation: !xmltv.applicable ? "M3U output requires XEPG mode." : !xmltv.ready ? "M3U output waits until every configured XMLTV source is ready." : mapping.activeCount == 0 ? "M3U output becomes usable after channels are active in Mapping." : m3uAvailable ? "Threadfin's generated channel playlist." : "The M3U output is not available yet."
      },
      {
        key: "xmltv",
        label: "XMLTV URL",
        value: xmltvAvailable ? overviewString(clientInfo["xepg-url"]) : "",
        available: xmltvAvailable,
        explanation: !xmltv.applicable ? "Guide data is managed by the client in PMS mode." : !xmltv.ready ? "XMLTV output waits until every configured guide source is ready." : mapping.activeCount == 0 ? "XMLTV output becomes usable after channels are active in Mapping." : xmltvAvailable ? "Threadfin's generated guide data." : "The XMLTV output is not available yet."
      }
    ]
  };
}
function selectActivityState(server) {
  var clientInfo = overviewRecord(overviewRecord(server).clientInfo);
  return {
    activeClientConnections: overviewNumber(clientInfo.activeClients),
    clients: {
      active: overviewNumber(clientInfo.activeClients),
      total: overviewNumber(clientInfo.totalClients)
    },
    playlists: {
      active: overviewNumber(clientInfo.activePlaylist),
      total: overviewNumber(clientInfo.totalPlaylist)
    }
  };
}
function selectAttentionState(server) {
  var root = overviewRecord(server);
  var log = overviewRecord(root.log);
  var clientInfo = overviewRecord(root.clientInfo);
  return {
    errors: overviewNumber(log.errors !== void 0 ? log.errors : clientInfo.errors),
    warnings: overviewNumber(log.warnings !== void 0 ? log.warnings : clientInfo.warnings)
  };
}
function selectSourceState(server) {
  var root = overviewRecord(server);
  var settings = overviewRecord(root.settings);
  var files = overviewRecord(settings.files);
  var sources = [];
  function appendSources(fileType, kind) {
    var configured = overviewRecord(files[fileType]);
    Object.keys(configured).forEach((key) => {
      var source = overviewRecord(configured[key]);
      var rawAvailability = source["provider.availability"];
      var availability = overviewNumber(rawAvailability);
      var ready = kind == "XMLTV" ? overviewXMLTVSourceReady(source) : availability > 0;
      sources.push({
        id: key,
        name: overviewString(source.name) || kind,
        kind,
        lastUpdate: overviewString(source["last.update"]),
        availability,
        status: rawAvailability === void 0 || rawAvailability === null ? "unknown" : ready ? "ready" : "unavailable"
      });
    });
  }
  appendSources("m3u", "Playlist");
  appendSources("hdhr", "HDHomeRun");
  appendSources("xmltv", "XMLTV");
  sources.sort((a, b) => {
    if (a.lastUpdate == b.lastUpdate) {
      return a.name.localeCompare(b.name);
    }
    return a.lastUpdate < b.lastUpdate ? 1 : -1;
  });
  return sources;
}
function selectOverviewStages(state) {
  var stages = [];
  stages.push(state.playlistCount == 0 ? {
    key: "playlist",
    label: "Playlist",
    status: "empty",
    summary: "No playlist configured",
    explanation: "Add an M3U playlist or HDHomeRun source to begin.",
    action: { label: "Add playlist", destination: "playlist", openAdd: "playlist" }
  } : state.playlistReadyCount == state.playlistCount ? {
    key: "playlist",
    label: "Playlist",
    status: "ready",
    summary: state.playlistReadyCount + " / " + state.playlistCount + " ready",
    explanation: "Every configured channel source is reachable.",
    action: { label: "View playlist", destination: "playlist" }
  } : {
    key: "playlist",
    label: "Playlist",
    status: "attention",
    summary: state.playlistReadyCount + " / " + state.playlistCount + " ready",
    explanation: "One or more configured channel sources need attention.",
    action: { label: "Review playlist", destination: "playlist" }
  });
  stages.push(state.playlistCount == 0 ? {
    key: "filter",
    label: "Filter",
    status: "waiting",
    summary: "Waiting for playlist",
    explanation: "Filters select which imported streams continue to Mapping.",
    action: { label: "View filters", destination: "filter" }
  } : state.selectedStreamCount > 0 ? {
    key: "filter",
    label: "Filter",
    status: "ready",
    summary: state.selectedStreamCount + " selected",
    explanation: "These imported streams continue through the signal path.",
    action: { label: "Review filters", destination: "filter" }
  } : {
    key: "filter",
    label: "Filter",
    status: "attention",
    summary: "No streams selected",
    explanation: "Create a filter to select imported streams for Mapping.",
    action: { label: "Create filter", destination: "filter" }
  });
  stages.push(!state.xmltv.applicable ? {
    key: "xmltv",
    label: "XMLTV",
    status: "managed",
    summary: "Managed by client",
    explanation: "PMS mode leaves guide data management to the connected client.",
    action: { label: "Review EPG source", destination: "settings" }
  } : state.xmltv.ready ? {
    key: "xmltv",
    label: "XMLTV",
    status: "ready",
    summary: state.xmltv.readyCount + " / " + state.xmltv.sourceCount + " ready",
    explanation: "Every configured guide source has channels available to XEPG.",
    action: { label: "View XMLTV", destination: "xmltv" }
  } : {
    key: "xmltv",
    label: "XMLTV",
    status: "attention",
    summary: state.xmltv.readyCount + " / " + state.xmltv.sourceCount + " ready",
    explanation: "XEPG needs every configured XMLTV source to be available with guide channels.",
    action: { label: "Add XMLTV", destination: "xmltv", openAdd: "xmltv" }
  });
  stages.push(!state.xmltv.applicable ? {
    key: "mapping",
    label: "Mapping",
    status: "managed",
    summary: "Managed by client",
    explanation: "PMS mode does not use Threadfin's XEPG Mapping workflow.",
    action: { label: "Review EPG source", destination: "settings" }
  } : state.selectedStreamCount == 0 ? {
    key: "mapping",
    label: "Mapping",
    status: "waiting",
    summary: "Waiting for selected streams",
    explanation: "Selected streams appear here after the Filter stage.",
    action: { label: "View Mapping", destination: "mapping" }
  } : state.mapping.unresolvedCount > 0 ? {
    key: "mapping",
    label: "Mapping",
    status: "attention",
    summary: state.mapping.unresolvedCount + " need attention",
    explanation: state.mapping.activeCount + " active; unresolved channels are not part of generated outputs.",
    action: { label: "Review " + state.mapping.unresolvedCount + " channels", destination: "mapping" }
  } : state.mapping.activeCount > 0 ? {
    key: "mapping",
    label: "Mapping",
    status: "ready",
    summary: state.mapping.activeCount + " active",
    explanation: "Active mappings are included in Threadfin's generated lineup.",
    action: { label: "View Mapping", destination: "mapping" }
  } : {
    key: "mapping",
    label: "Mapping",
    status: "attention",
    summary: "No active mappings",
    explanation: "Activate and map channels before using XEPG outputs.",
    action: { label: "Review Mapping", destination: "mapping" }
  });
  stages.push(state.outputs.ready ? {
    key: "outputs",
    label: "Outputs",
    status: "ready",
    summary: "Threadfin output ready",
    explanation: "Endpoints are ready to copy into a supported client; client setup is not observable here.",
    action: { label: "View output endpoints", destination: "connections" }
  } : {
    key: "outputs",
    label: "Outputs",
    status: "waiting",
    summary: "Waiting for lineup",
    explanation: "Complete the applicable stages before connecting a client.",
    action: { label: "View Connections", destination: "connections" }
  });
  return stages;
}
function selectOverviewState(server) {
  var playlistCount = selectPlaylistCount(server);
  var selectedStreamCount = selectSelectedStreamCount(server);
  var xmltv = selectXMLTVState(server);
  var mapping = selectMappingState(server);
  var outputs = selectOutputState(server, selectedStreamCount, xmltv, mapping);
  var sources = selectSourceState(server);
  var playlistReadyCount = sources.filter((source) => source.kind != "XMLTV" && source.status == "ready").length;
  var stateWithoutStages = {
    playlistCount,
    playlistReadyCount,
    selectedStreamCount,
    xmltv,
    mapping,
    outputs,
    activity: selectActivityState(server),
    attention: selectAttentionState(server),
    sources
  };
  var state = stateWithoutStages;
  state.stages = selectOverviewStages(stateWithoutStages);
  return state;
}
function renderOverview(server) {
  var host = document.getElementById("overview-content");
  if (!host) {
    return;
  }
  var state = selectOverviewState(server);
  while (host.firstChild) {
    host.removeChild(host.firstChild);
  }
  var overview = document.createElement("div");
  overview.className = "tf-overview";
  var header = document.createElement("header");
  header.className = "tf-overview-header";
  var title = document.createElement("h1");
  title.textContent = "Overview";
  var introduction = document.createElement("p");
  introduction.textContent = "Follow the Threadfin signal path from imported sources to generated outputs.";
  header.appendChild(title);
  header.appendChild(introduction);
  overview.appendChild(header);
  var pathHeading = document.createElement("h2");
  pathHeading.textContent = "Signal path";
  overview.appendChild(pathHeading);
  overview.appendChild(createOverviewSignalPath(state.stages));
  var panels = document.createElement("div");
  panels.className = "tf-overview-panels";
  panels.appendChild(createOverviewActivityPanel(state));
  panels.appendChild(createOverviewAttentionPanel(state));
  panels.appendChild(createOverviewEndpointsPanel(state.outputs.endpoints));
  panels.appendChild(createOverviewSourcesPanel(state.sources));
  overview.appendChild(panels);
  var copyStatus = document.createElement("p");
  copyStatus.id = "overview-copy-status";
  copyStatus.className = "tf-overview-status";
  copyStatus.setAttribute("role", "status");
  copyStatus.setAttribute("aria-live", "polite");
  copyStatus.setAttribute("aria-atomic", "true");
  overview.appendChild(copyStatus);
  host.appendChild(overview);
}
function createOverviewSignalPath(stages) {
  var path = document.createElement("ol");
  path.className = "tf-signal-path";
  path.setAttribute("aria-label", "Threadfin signal path");
  stages.forEach((stage) => {
    var item = document.createElement("li");
    item.className = "tf-signal-stage";
    item.setAttribute("data-stage", stage.key);
    item.setAttribute("data-status", stage.status);
    var status = document.createElement("p");
    status.className = "tf-signal-stage-status";
    status.textContent = overviewStatusLabel(stage.status);
    var heading = document.createElement("h3");
    heading.textContent = stage.label;
    var summary = document.createElement("p");
    summary.className = "tf-signal-stage-summary";
    summary.textContent = stage.summary;
    var explanation = document.createElement("p");
    explanation.className = "tf-signal-stage-explanation";
    explanation.textContent = stage.explanation;
    var action = document.createElement("button");
    action.type = "button";
    action.className = "tf-overview-action";
    action.textContent = stage.action.label;
    action.addEventListener("click", function() {
      openDestination(stage.action.destination, true);
      if (stage.action.openAdd && typeof openSourcePopup == "function") {
        var addButton = document.querySelector('[data-source-focus-key="' + stage.action.openAdd + ':add"]');
        if (addButton) {
          openSourcePopup(stage.action.openAdd, void 0, addButton);
        }
      }
    });
    item.appendChild(status);
    item.appendChild(heading);
    item.appendChild(summary);
    item.appendChild(explanation);
    item.appendChild(action);
    path.appendChild(item);
  });
  return path;
}
function overviewStatusLabel(status) {
  switch (status) {
    case "ready":
      return "Ready";
    case "attention":
      return "Needs attention";
    case "empty":
      return "Not configured";
    case "managed":
      return "Managed by client";
    default:
      return "Waiting";
  }
}
function createOverviewPanel(title, className) {
  var panel = document.createElement("section");
  panel.className = "tf-overview-panel " + className;
  var heading = document.createElement("h2");
  heading.textContent = title;
  panel.appendChild(heading);
  return panel;
}
function createOverviewActivityPanel(state) {
  var panel = createOverviewPanel("Current activity", "tf-overview-activity");
  var metrics = document.createElement("dl");
  metrics.className = "tf-overview-metrics";
  appendOverviewMetric(metrics, "Active client connections", "overview-active-client-connections", String(state.activity.activeClientConnections));
  appendOverviewMetric(metrics, "Client capacity", "overview-client-capacity", overviewCapacityLabel(state.activity.clients));
  appendOverviewMetric(metrics, "Playlist connections", "overview-playlist-capacity", overviewCapacityLabel(state.activity.playlists));
  panel.appendChild(metrics);
  var explanation = document.createElement("p");
  explanation.className = "tf-overview-panel-note";
  explanation.id = "overview-activity-summary";
  explanation.textContent = overviewActivitySummary(state.activity.activeClientConnections);
  panel.appendChild(explanation);
  return panel;
}
function overviewActivitySummary(activeClientConnections) {
  if (activeClientConnections == 0) {
    return "No active client connections are currently observed.";
  }
  return "Threadfin currently observes " + activeClientConnections + (activeClientConnections == 1 ? " active client connection." : " active client connections.");
}
function appendOverviewMetric(list, label, id, value) {
  var item = document.createElement("div");
  var term = document.createElement("dt");
  term.textContent = label;
  var description = document.createElement("dd");
  description.id = id;
  description.textContent = value;
  item.appendChild(term);
  item.appendChild(description);
  list.appendChild(item);
}
function overviewCapacityLabel(capacity) {
  return capacity.active + " / " + capacity.total;
}
function createOverviewAttentionPanel(state) {
  var panel = createOverviewPanel("Attention required", "tf-overview-attention");
  var counts = document.createElement("dl");
  counts.className = "tf-overview-metrics";
  appendOverviewMetric(counts, "Errors", "overview-errors", String(state.attention.errors));
  appendOverviewMetric(counts, "Warnings", "overview-warnings", String(state.attention.warnings));
  panel.appendChild(counts);
  var explanation = document.createElement("p");
  explanation.className = "tf-overview-panel-note";
  explanation.id = "overview-attention-summary";
  explanation.textContent = overviewAttentionSummary(state.attention);
  panel.appendChild(explanation);
  var action = document.createElement("button");
  action.type = "button";
  action.className = "tf-overview-action";
  action.textContent = "View Log";
  action.addEventListener("click", function() {
    openDestination("log", true);
  });
  panel.appendChild(action);
  return panel;
}
function overviewAttentionSummary(attention) {
  if (attention.errors == 0 && attention.warnings == 0) {
    return "Threadfin reports no current errors or warnings.";
  }
  return "Threadfin reports " + attention.errors + (attention.errors == 1 ? " error and " : " errors and ") + attention.warnings + (attention.warnings == 1 ? " warning." : " warnings.");
}
function createOverviewEndpointsPanel(endpoints) {
  var panel = createOverviewPanel("Output endpoints", "tf-overview-endpoints");
  var list = document.createElement("ul");
  list.className = "tf-endpoint-list";
  endpoints.forEach((endpoint) => {
    var item = document.createElement("li");
    item.className = "tf-endpoint";
    var heading = document.createElement("h3");
    heading.textContent = endpoint.label;
    var value = document.createElement("code");
    value.textContent = endpoint.available ? endpoint.value : "Unavailable";
    var explanation = document.createElement("p");
    explanation.textContent = endpoint.explanation;
    item.appendChild(heading);
    item.appendChild(value);
    item.appendChild(explanation);
    if (endpoint.available) {
      var copy = document.createElement("button");
      copy.type = "button";
      copy.className = "tf-overview-action tf-copy-action";
      copy.textContent = "Copy " + endpoint.label;
      bindOverviewCopyAction(copy, endpoint.value, endpoint.label);
      item.appendChild(copy);
    }
    list.appendChild(item);
  });
  panel.appendChild(list);
  return panel;
}
function bindOverviewCopyAction(button, value, label, statusID = "overview-copy-status") {
  if (typeof import_clipboard.default != "undefined") {
    var helper = new import_clipboard.default(button, {
      text: function() {
        return value;
      }
    });
    helper.on("success", function(event) {
      if (event && event.clearSelection) {
        event.clearSelection();
      }
      announceOverviewCopyStatus(label + " copied.", statusID);
    });
    helper.on("error", function() {
      copyOverviewWithFallback(value, label, statusID);
    });
    return;
  }
  button.addEventListener("click", function() {
    copyOverviewWithFallback(value, label, statusID);
  });
}
function copyOverviewWithFallback(value, label, statusID = "overview-copy-status") {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(value).then(function() {
      announceOverviewCopyStatus(label + " copied.", statusID);
    }).catch(function() {
      copyOverviewWithSelection(value, label, statusID);
    });
    return;
  }
  copyOverviewWithSelection(value, label, statusID);
}
function copyOverviewWithSelection(value, label, statusID = "overview-copy-status") {
  var temporary = document.createElement("textarea");
  temporary.value = value;
  temporary.setAttribute("readonly", "");
  temporary.style.position = "fixed";
  temporary.style.opacity = "0";
  document.body.appendChild(temporary);
  temporary.select();
  var copied = false;
  try {
    copied = document.execCommand("copy");
  } catch (_) {
    copied = false;
  }
  document.body.removeChild(temporary);
  announceOverviewCopyStatus(copied ? label + " copied." : "Copy failed. Select and copy the endpoint manually.", statusID);
}
function announceOverviewCopyStatus(message, statusID = "overview-copy-status") {
  var status = document.getElementById(statusID);
  if (status) {
    status.textContent = message;
  }
}
function createOverviewSourcesPanel(sources) {
  var panel = createOverviewPanel("Recent source state", "tf-overview-sources");
  if (sources.length == 0) {
    var empty = document.createElement("p");
    empty.className = "tf-overview-panel-note";
    empty.textContent = "No Playlist or XMLTV sources are configured yet.";
    panel.appendChild(empty);
    return panel;
  }
  var list = document.createElement("ul");
  list.className = "tf-source-state-list";
  sources.forEach((source) => {
    var item = document.createElement("li");
    item.className = "tf-source-state";
    item.setAttribute("data-status", source.status);
    var heading = document.createElement("h3");
    heading.textContent = source.name;
    var kind = document.createElement("p");
    kind.className = "tf-source-state-kind";
    kind.textContent = source.kind + " \xB7 " + overviewSourceStatusLabel(source);
    var updated = document.createElement("p");
    updated.textContent = source.lastUpdate ? "Last update: " + source.lastUpdate : "No successful update recorded";
    item.appendChild(heading);
    item.appendChild(kind);
    item.appendChild(updated);
    list.appendChild(item);
  });
  panel.appendChild(list);
  return panel;
}
function overviewSourceStatusLabel(source) {
  if (source.status == "unknown") {
    return "Availability unknown";
  }
  if (source.status == "unavailable") {
    return "Unavailable";
  }
  return source.availability + "% available";
}
function refreshOverviewOperationalState(response) {
  var activity = selectActivityState(response);
  var attention = selectAttentionState(response);
  setOverviewOperationalText("overview-active-client-connections", String(activity.activeClientConnections));
  setOverviewOperationalText("overview-client-capacity", overviewCapacityLabel(activity.clients));
  setOverviewOperationalText("overview-playlist-capacity", overviewCapacityLabel(activity.playlists));
  setOverviewOperationalText("overview-activity-summary", overviewActivitySummary(activity.activeClientConnections));
  setOverviewOperationalText("overview-errors", String(attention.errors));
  setOverviewOperationalText("overview-warnings", String(attention.warnings));
  setOverviewOperationalText("overview-attention-summary", overviewAttentionSummary(attention));
}
function setOverviewOperationalText(id, value) {
  var element = document.getElementById(id);
  if (element && element.textContent != value) {
    element.textContent = value;
  }
}
function renderConnections(server) {
  var host = document.getElementById("connections-content");
  if (!host) {
    return;
  }
  var endpoints = selectOverviewState(server).outputs.endpoints;
  while (host.firstChild) {
    host.removeChild(host.firstChild);
  }
  var connections = document.createElement("div");
  connections.className = "tf-connections";
  var header = document.createElement("header");
  header.className = "tf-operations-header";
  var title = document.createElement("h1");
  title.textContent = "Connections";
  var introduction = document.createElement("p");
  introduction.textContent = "Copy an available Threadfin endpoint into a supported client. Client setup and success are verified in that client, not by Threadfin.";
  header.appendChild(title);
  header.appendChild(introduction);
  connections.appendChild(header);
  var guidance = document.createElement("section");
  guidance.className = "tf-operations-panel tf-connections-guidance";
  var guidanceHeading = document.createElement("h2");
  guidanceHeading.textContent = "Choose an endpoint";
  var guidanceText = document.createElement("p");
  guidanceText.textContent = "Use the DVR address for tuner discovery. Use M3U for the generated channel playlist and XMLTV for generated guide data when those outputs are available.";
  guidance.appendChild(guidanceHeading);
  guidance.appendChild(guidanceText);
  connections.appendChild(guidance);
  var endpointSection = document.createElement("section");
  endpointSection.className = "tf-operations-panel";
  var endpointHeading = document.createElement("h2");
  endpointHeading.textContent = "Threadfin endpoints";
  endpointSection.appendChild(endpointHeading);
  var endpointList = document.createElement("ul");
  endpointList.className = "tf-connections-endpoints";
  endpoints.forEach((endpoint) => endpointList.appendChild(createConnectionsEndpoint(endpoint)));
  endpointSection.appendChild(endpointList);
  connections.appendChild(endpointSection);
  var clientNotes = document.createElement("section");
  clientNotes.className = "tf-operations-panel tf-client-notes";
  var clientNotesHeading = document.createElement("h2");
  clientNotesHeading.textContent = "Client-specific notes";
  var clientNotesIntroduction = document.createElement("p");
  clientNotesIntroduction.textContent = "These notes identify where to start. Complete and verify configuration in the selected client.";
  clientNotes.appendChild(clientNotesHeading);
  clientNotes.appendChild(clientNotesIntroduction);
  clientNotes.appendChild(createConnectionsClientNote("Plex", "Use the DVR address when adding Threadfin as a tuner. Complete the remaining guide and channel setup in Plex."));
  clientNotes.appendChild(createConnectionsClientNote("Jellyfin", "Choose the DVR address or the available M3U and XMLTV endpoints for the setup flow supported by your Jellyfin version. Verify the result in Jellyfin."));
  clientNotes.appendChild(createConnectionsClientNote("Emby", "Choose the DVR address or the available M3U and XMLTV endpoints for the setup flow supported by your Emby version. Verify the result in Emby."));
  connections.appendChild(clientNotes);
  var copyStatus = document.createElement("p");
  copyStatus.id = "connections-copy-status";
  copyStatus.className = "tf-overview-status";
  copyStatus.setAttribute("role", "status");
  copyStatus.setAttribute("aria-live", "polite");
  copyStatus.setAttribute("aria-atomic", "true");
  connections.appendChild(copyStatus);
  host.appendChild(connections);
}
function createConnectionsEndpoint(endpoint) {
  var item = document.createElement("li");
  item.className = "tf-connection-endpoint";
  item.setAttribute("data-endpoint", endpoint.key);
  item.setAttribute("data-available", String(endpoint.available));
  var heading = document.createElement("h3");
  heading.textContent = endpoint.label;
  item.appendChild(heading);
  if (endpoint.available) {
    var value = document.createElement("code");
    value.className = "tf-endpoint-value";
    value.textContent = endpoint.value;
    item.appendChild(value);
    var explanation = document.createElement("p");
    explanation.textContent = endpoint.explanation;
    item.appendChild(explanation);
    var copy = document.createElement("button");
    copy.type = "button";
    copy.className = "tf-overview-action tf-copy-action";
    copy.textContent = "Copy " + endpoint.label;
    bindOverviewCopyAction(copy, endpoint.value, endpoint.label, "connections-copy-status");
    item.appendChild(copy);
  } else {
    var unavailable = document.createElement("p");
    unavailable.className = "tf-endpoint-unavailable";
    unavailable.textContent = "Unavailable \u2014 " + endpoint.explanation;
    item.appendChild(unavailable);
  }
  return item;
}
function createConnectionsClientNote(client, note) {
  var details = document.createElement("details");
  var summary = document.createElement("summary");
  summary.textContent = client;
  var explanation = document.createElement("p");
  explanation.textContent = note;
  details.appendChild(summary);
  details.appendChild(explanation);
  return details;
}
function renderActivity(server) {
  refreshActivityOperationalState(server);
  if (activityHasStreamPreview(server) && typeof showPreview == "function") {
    showPreview(true);
  }
}
function refreshActivityOperationalState(server) {
  var activity = selectActivityState(server);
  setActivityText("playlist-connection-information", activityCapacityLabel("Playlist source connections", activity.playlists));
  setActivityText("client-connection-information", activityCapacityLabel("Client connections", activity.clients));
}
function activityCapacityLabel(label, capacity) {
  return label + ": " + capacity.active + " / " + capacity.total;
}
function setActivityText(id, value) {
  var element = document.getElementById(id);
  if (element && element.textContent != value) {
    element.textContent = value;
  }
}
function activityHasStreamPreview(server) {
  var root = overviewRecord(server);
  var data = overviewRecord(root.data);
  var preview = overviewRecord(data.StreamPreviewUI);
  return Array.isArray(preview.activeStreams) && Array.isArray(preview.inactiveStreams);
}
var navigationGroups = [
  { key: "overview", label: "Overview", items: ["overview"] },
  { key: "sources", label: "Sources", items: ["playlist", "xmltv"] },
  { key: "lineup", label: "Lineup", items: ["filter", "mapping"] },
  { key: "delivery", label: "Delivery", items: ["connections", "users"] },
  { key: "system", label: "System", items: ["activity", "settings", "log"] },
  { key: "account", label: "Account", items: ["logout"] }
];
var legacyMenuIndexByDestination = {
  playlist: 0,
  xmltv: 1,
  filter: 2,
  mapping: 3,
  users: 4,
  settings: 5,
  log: 6,
  logout: 7
};
var legacyDestinationByMenuIndex = {
  0: "playlist",
  1: "xmltv",
  2: "filter",
  3: "mapping",
  4: "users",
  5: "settings",
  6: "log",
  7: "logout"
};
var currentDestination;
var initialDestinationRestored = false;
var navigationGuardBypass = false;
var sidebarNavigationStorageKey = "threadfin.navigation.collapsed";
var navigationIconByDestination = {
  overview: "fa-tachometer-alt",
  playlist: "fa-list",
  xmltv: "fa-calendar-alt",
  filter: "fa-filter",
  mapping: "fa-random",
  connections: "fa-plug",
  users: "fa-users",
  activity: "fa-broadcast-tower",
  settings: "fa-cog",
  log: "fa-file-alt",
  logout: "fa-sign-out-alt"
};
function renderNavigation() {
  var navigation = document.getElementById("main-menu");
  if (!navigation) {
    return;
  }
  var navigationElement = navigation;
  navigationElement.innerHTML = "";
  navigationGroups.forEach((group) => {
    var items = group.items.filter((destination) => navigationDestinationIsVisible(destination));
    if (items.length == 0) {
      return;
    }
    var groupElement = document.createElement("section");
    groupElement.className = "tf-navigation-group";
    groupElement.setAttribute("aria-labelledby", "navigation-group-" + group.key);
    var heading = document.createElement("h2");
    heading.id = "navigation-group-" + group.key;
    heading.className = "tf-navigation-group-label";
    heading.textContent = group.label;
    groupElement.appendChild(heading);
    var list = document.createElement("ul");
    list.className = "tf-navigation-list";
    items.forEach((destination) => {
      var listItem = document.createElement("li");
      var button = document.createElement("button");
      button.type = "button";
      button.className = "tf-navigation-item";
      button.setAttribute("data-destination", destination);
      appendNavigationItemContent(button, navigationIconByDestination[destination], navigationDestinationLabel(destination));
      button.addEventListener("click", function() {
        openDestination(destination, true, button);
      });
      listItem.appendChild(button);
      list.appendChild(listItem);
    });
    if (group.key == "system") {
      var documentationItem = document.createElement("li");
      var documentationLink = document.createElement("a");
      documentationLink.className = "tf-navigation-item";
      documentationLink.setAttribute("href", "https://github.com/volschin/Threadfin/blob/main/docs/user-guide.md");
      documentationLink.setAttribute("target", "_blank");
      documentationLink.setAttribute("rel", "noopener noreferrer");
      appendNavigationItemContent(documentationLink, "fa-book-open", "User guide", true);
      documentationItem.appendChild(documentationLink);
      list.appendChild(documentationItem);
    }
    groupElement.appendChild(list);
    navigationElement.appendChild(groupElement);
  });
  renderLegacyMenuAdapters(navigationElement);
  updateNavigationCurrentPage();
  initializeSidebarNavigation();
}
function appendNavigationItemContent(element, iconName, label, external = false) {
  element.setAttribute("aria-label", label);
  element.setAttribute("title", label);
  var icon = document.createElement("i");
  icon.className = "fas " + iconName + " tf-navigation-icon";
  icon.setAttribute("aria-hidden", "true");
  element.appendChild(icon);
  var text = document.createElement("span");
  text.className = "tf-navigation-label";
  text.textContent = label;
  element.appendChild(text);
  if (external) {
    var externalIcon = document.createElement("i");
    externalIcon.className = "fas fa-external-link-alt tf-navigation-external";
    externalIcon.setAttribute("aria-hidden", "true");
    element.appendChild(externalIcon);
  }
}
function initializeSidebarNavigation() {
  var app = document.querySelector(".tf-app");
  var toggle = document.querySelector(".tf-sidebar-toggle");
  if (!app || !toggle) {
    return;
  }
  var collapsed = false;
  try {
    collapsed = localStorage.getItem(sidebarNavigationStorageKey) == "true";
  } catch (error) {
    console.warn("Navigation preference is unavailable", error);
  }
  applySidebarNavigationState(app, toggle, collapsed);
  if (toggle.getAttribute("data-sidebar-toggle-bound") == "true") {
    return;
  }
  toggle.setAttribute("data-sidebar-toggle-bound", "true");
  toggle.addEventListener("click", function() {
    var nextCollapsed = !app.classList.contains("tf-sidebar-collapsed");
    applySidebarNavigationState(app, toggle, nextCollapsed);
    try {
      localStorage.setItem(sidebarNavigationStorageKey, String(nextCollapsed));
    } catch (error) {
      console.warn("Navigation preference could not be saved", error);
    }
  });
}
function applySidebarNavigationState(app, toggle, collapsed) {
  app.classList.toggle("tf-sidebar-collapsed", collapsed);
  toggle.setAttribute("aria-expanded", String(!collapsed));
  var label = collapsed ? "Expand navigation" : "Collapse navigation";
  toggle.setAttribute("aria-label", label);
  toggle.setAttribute("title", label);
}
function renderLegacyMenuAdapters(navigation) {
  var adapters = document.createElement("div");
  adapters.id = "legacy-menu-adapters";
  adapters.hidden = true;
  Object.keys(legacyDestinationByMenuIndex).forEach((key) => {
    var index = Number(key);
    var adapter = document.createElement("button");
    adapter.type = "button";
    adapter.id = key;
    adapter.setAttribute("data-legacy-menu-index", key);
    adapter.addEventListener("click", function() {
      openLegacyMenu(index, true, adapter);
    });
    adapters.appendChild(adapter);
  });
  navigation.appendChild(adapters);
}
function navigationDestinationIsVisible(destination) {
  if (destination != "users" && destination != "logout") {
    return true;
  }
  var server = SERVER;
  return server["settings"] && server["settings"]["authentication.web"] == true;
}
function navigationDestinationLabel(destination) {
  switch (destination) {
    case "overview":
      return "Overview";
    case "connections":
      return "Connections";
    case "activity":
      return "Activity";
    default:
      var index = legacyMenuIndexByDestination[destination];
      return menuItems[index].value;
  }
}
function openDestination(destination, addHistory, invoker) {
  if (guardMappingDestination(destination, function() {
    openDestination(destination, addHistory, invoker);
  }, invoker)) {
    return;
  }
  var legacyIndex = legacyMenuIndexByDestination[destination];
  if (legacyIndex !== void 0) {
    navigationGuardBypass = true;
    try {
      openLegacyMenu(legacyIndex, false);
    } finally {
      navigationGuardBypass = false;
    }
  } else {
    showDestinationHost(destination);
    if (destination == "overview") {
      renderOverview(SERVER);
    } else if (destination == "connections") {
      renderConnections(SERVER);
    } else if (destination == "activity") {
      renderActivity(SERVER);
    }
    setCurrentDestination(destination);
    dismissMobileNavigation();
    focusMainContent();
  }
  if (addHistory) {
    window.history.pushState({ threadfinDestination: destination }, "", "#" + destination);
  }
}
function openLegacyMenu(index, replaceHistory = true, invoker) {
  var destination = legacyDestinationByMenuIndex[index];
  if (!destination) {
    console.warn("Ignoring unknown legacy menu index", index);
    return;
  }
  if (guardMappingDestination(destination, function() {
    openLegacyMenu(index, replaceHistory, invoker);
  }, invoker)) {
    return;
  }
  showDestinationHost("content");
  var content = new ShowContent(index);
  content.show();
  setCurrentDestination(destination);
  dismissMobileNavigation();
  focusMainContent();
  if (replaceHistory) {
    window.history.replaceState({ threadfinDestination: destination }, "", "#" + destination);
  }
}
function guardMappingDestination(destination, continuation, invoker) {
  if (navigationGuardBypass || currentDestination != "mapping" || destination == "mapping" || typeof mappingHasDirtyDraft != "function" || !mappingHasDirtyDraft() || typeof mappingRequestNavigation != "function") {
    return false;
  }
  mappingRequestNavigation(function() {
    navigationGuardBypass = true;
    try {
      continuation();
    } finally {
      navigationGuardBypass = false;
    }
  }, invoker);
  return true;
}
function showDestinationHost(destination) {
  var hostByDestination = {
    overview: "overview-content",
    connections: "connections-content",
    activity: "activity-content",
    content: "content"
  };
  Object.keys(hostByDestination).forEach((key) => {
    var host = document.getElementById(hostByDestination[key]);
    if (host) {
      host.hidden = key != destination;
    }
  });
}
function setCurrentDestination(destination) {
  currentDestination = destination;
  updateNavigationCurrentPage();
}
function updateNavigationCurrentPage() {
  var items = document.querySelectorAll("#main-menu [data-destination]");
  for (var i = 0; i < items.length; i++) {
    var item = items[i];
    if (item.getAttribute("data-destination") == currentDestination) {
      item.setAttribute("aria-current", "page");
    } else {
      item.removeAttribute("aria-current");
    }
  }
}
function dismissMobileNavigation() {
  var navigation = document.getElementById("main-menu");
  var toggle = document.querySelector(".tf-nav-toggle");
  if (!navigation) {
    return;
  }
  var drawer = Drawer.getInstance(document.getElementById("navigation-drawer"));
  if (drawer)
    drawer.hide();
  if (toggle) {
    toggle.setAttribute("aria-expanded", "false");
  }
}
function focusMainContent() {
  var main = document.getElementById("main-content");
  if (main) {
    window.setTimeout(function() {
      main.focus({ preventScroll: true });
    }, 0);
  }
}
function restoreDestinationFromHistory() {
  var destination = window.history.state && window.history.state.threadfinDestination;
  if (!destination && window.location.hash.length > 1) {
    destination = window.location.hash.slice(1);
  }
  if (navigationDestinationIsKnown(destination) && navigationDestinationIsVisible(destination)) {
    if (currentDestination == "mapping" && destination != "mapping" && typeof mappingHasDirtyDraft == "function" && mappingHasDirtyDraft() && typeof mappingRequestNavigation == "function") {
      window.history.replaceState({ threadfinDestination: "mapping" }, "", "#mapping");
      mappingRequestNavigation(function() {
        navigationGuardBypass = true;
        try {
          openDestination(destination, false);
          window.history.replaceState({ threadfinDestination: destination }, "", "#" + destination);
        } finally {
          navigationGuardBypass = false;
        }
      }, void 0);
      return true;
    }
    openDestination(destination, false);
    return true;
  }
  return false;
}
function restoreInitialDestinationFromHistory() {
  if (initialDestinationRestored || currentDestination !== void 0) {
    return;
  }
  initialDestinationRestored = true;
  if (!restoreDestinationFromHistory()) {
    openDestination("overview", false);
    window.history.replaceState({ threadfinDestination: "overview" }, "", "#overview");
  }
}
function navigationDestinationIsKnown(destination) {
  return navigationGroups.some((group) => group.items.indexOf(destination) != -1);
}
window.addEventListener("popstate", restoreDestinationFromHistory);
var SettingsCategory = class {
  constructor() {
    this.DocumentID = "content_settings";
  }
  createCategoryHeadline(value) {
    var element = document.createElement("H4");
    element.innerHTML = value;
    return element;
  }
  createHR() {
    var element = document.createElement("HR");
    return element;
  }
  createSettings(settingsKey) {
    var setting = document.createElement("TR");
    var content = new PopupContent();
    var data = SERVER["settings"][settingsKey];
    switch (settingsKey) {
      // Texteingaben
      case "update":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.update.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "update", data.toString());
        input.setAttribute("placeholder", "{{.settings.update.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "backup.path":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.backupPath.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "backup.path", data);
        input.setAttribute("placeholder", "{{.settings.backupPath.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "temp.path":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.tempPath.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "temp.path", data);
        input.setAttribute("placeholder", "{{.settings.tmpPath.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "user.agent":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.userAgent.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "user.agent", data);
        input.setAttribute("placeholder", "{{.settings.userAgent.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "buffer.timeout":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.bufferTimeout.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "buffer.timeout", data);
        input.setAttribute("placeholder", "{{.settings.bufferTimeout.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "ffmpeg.path":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.ffmpegPath.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "ffmpeg.path", data);
        input.setAttribute("placeholder", "{{.settings.ffmpegPath.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "ffmpeg.options":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.ffmpegOptions.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "ffmpeg.options", data);
        input.setAttribute("placeholder", "{{.settings.ffmpegOptions.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "ffmpeg.forceHttp":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.ffmpegForceHttp.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "vlc.path":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.vlcPath.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "vlc.path", data);
        input.setAttribute("placeholder", "{{.settings.vlcPath.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "vlc.options":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.vlcOptions.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "vlc.options", data);
        input.setAttribute("placeholder", "{{.settings.vlcOptions.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      // Checkboxen
      case "authentication.web":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.authenticationWEB.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "authentication.pms":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.authenticationPMS.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "authentication.m3u":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.authenticationM3U.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "authentication.xml":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.authenticationXML.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "authentication.api":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.authenticationAPI.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "files.update":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.filesUpdate.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "cache.images":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.cacheImages.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "xepg.replace.missing.images":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.replaceEmptyImages.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "xepg.replace.channel.title":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.replaceChannelTitle.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "storeBufferInRAM":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.storeBufferInRAM.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "forceHttps":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.forceHttps.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "excludeStreamHttps":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.excludeStreamHttps.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "httpsPort":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.httpsPort.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "httpsPort", data.toString());
        input.setAttribute("placeholder", "{{.settings.httpsPort.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "httpsThreadfinDomain":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.httpsThreadfinDomain.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "httpsThreadfinDomain", data.toString());
        input.setAttribute("placeholder", "{{.settings.httpsThreadfinDomain.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "bindIpAddress":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.bindIpAddress.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "bindIpAddress", data.toString());
        input.setAttribute("placeholder", "{{.settings.bindIpAddress.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "httpThreadfinDomain":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.httpThreadfinDomain.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "httpThreadfinDomain", data.toString());
        input.setAttribute("placeholder", "{{.settings.httpThreadfinDomain.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "enableNonAscii":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.enableNonAscii.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "epgCategories":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.epgCategories.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "epgCategories", data.toString());
        input.setAttribute("placeholder", "{{.settings.epgCategories.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "epgCategoriesColors":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.epgCategoriesColors.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "epgCategoriesColors", data.toString());
        input.setAttribute("placeholder", "{{.settings.epgCategoriesColors.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "ThreadfinAutoUpdate":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.ThreadfinAutoUpdate.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "ssdp":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.ssdp.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "dummy":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.dummy.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "dummyChannel":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.dummyChannel.title}}:";
        var tdRight = document.createElement("TD");
        var text = ["PPV", "30 Minutes", "60 Minutes", "90 Minutes", "120 Minutes", "180 Minutes", "240 Minutes", "360 Minutes"];
        var values = ["PPV", "30_Minutes", "60_Minutes", "90_Minutes", "120_Minutes", "180_Minutes", "240_Minutes", "360_Minutes"];
        var select = content.createSelect(text, values, data, settingsKey);
        addRuntimeEventListener(select, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(select);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "ignoreFilters":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.ignoreFilters.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "api":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.api.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createCheckbox(settingsKey);
        input.checked = data;
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      // Select
      case "tuner":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.tuner.title}}:";
        var tdRight = document.createElement("TD");
        var text = new Array();
        var values = new Array();
        for (var i = 1; i <= 100; i++) {
          text.push(i);
          values.push(i);
        }
        var select = content.createSelect(text, values, data, settingsKey);
        addRuntimeEventListener(select, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(select);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "epgSource":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.epgSource.title}}:";
        var tdRight = document.createElement("TD");
        var text = ["PMS", "XEPG"];
        var values = ["PMS", "XEPG"];
        var select = content.createSelect(text, values, data, settingsKey);
        addRuntimeEventListener(select, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(select);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "backup.keep":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.backupKeep.title}}:";
        var tdRight = document.createElement("TD");
        var text = ["5", "10", "20", "30", "40", "50"];
        var values = ["5", "10", "20", "30", "40", "50"];
        var select = content.createSelect(text, values, data, settingsKey);
        addRuntimeEventListener(select, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(select);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "buffer.size.kb":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.bufferSize.title}}:";
        var tdRight = document.createElement("TD");
        var text = ["0.5 MB", "1 MB", "2 MB", "3 MB", "4 MB", "5 MB", "6 MB", "7 MB", "8 MB"];
        var values = ["512", "1024", "2048", "3072", "4096", "5120", "6144", "7168", "8192"];
        var select = content.createSelect(text, values, data, settingsKey);
        addRuntimeEventListener(select, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(select);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "buffer":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.streamBuffering.title}}:";
        var tdRight = document.createElement("TD");
        var text = ["{{.settings.streamBuffering.info_false}}", "FFmpeg: ({{.settings.streamBuffering.info_ffmpeg}})", "VLC: ({{.settings.streamBuffering.info_vlc}})"];
        var values = ["-", "ffmpeg", "vlc"];
        var select = content.createSelect(text, values, data, settingsKey);
        addRuntimeEventListener(select, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(select);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
      case "udpxy":
        var tdLeft = document.createElement("TD");
        tdLeft.innerHTML = "{{.settings.udpxy.title}}:";
        var tdRight = document.createElement("TD");
        var input = content.createInput("text", "udpxy", data);
        input.setAttribute("placeholder", "{{.settings.udpxy.placeholder}}");
        addRuntimeEventListener(input, "change", (event) => {
          event.currentTarget.className = "changed";
        });
        tdRight.appendChild(input);
        setting.appendChild(tdLeft);
        setting.appendChild(tdRight);
        break;
    }
    return setting;
  }
  createDescription(settingsKey) {
    var description = document.createElement("TR");
    var text;
    switch (settingsKey) {
      case "authentication.web":
        text = "{{.settings.authenticationWEB.description}}";
        break;
      case "authentication.m3u":
        text = "{{.settings.authenticationM3U.description}}";
        break;
      case "authentication.pms":
        text = "{{.settings.authenticationPMS.description}}";
        break;
      case "authentication.xml":
        text = "{{.settings.authenticationXML.description}}";
        break;
      case "authentication.api":
        if (SERVER["settings"]["authentication.web"] == true) {
          text = "{{.settings.authenticationAPI.description}}";
        }
        break;
      case "ThreadfinAutoUpdate":
        text = "{{.settings.ThreadfinAutoUpdate.description}}";
        break;
      case "backup.keep":
        text = "{{.settings.backupKeep.description}}";
        break;
      case "backup.path":
        text = "{{.settings.backupPath.description}}";
        break;
      case "temp.path":
        text = "{{.settings.tempPath.description}}";
        break;
      case "buffer":
        text = "{{.settings.streamBuffering.description}}";
        break;
      case "buffer.size.kb":
        text = "{{.settings.bufferSize.description}}";
        break;
      case "storeBufferInRAM":
        text = "{{.settings.storeBufferInRAM.description}}";
        break;
      case "forceHttps":
        text = "{{.settings.forceHttps.description}}";
        break;
      case "excludeStreamHttps":
        text = "{{.settings.excludeStreamHttps.description}}";
        break;
      case "httpsPort":
        text = "{{.settings.httpsPort.description}}";
        break;
      case "httpsThreadfinDomain":
        text = "{{.settings.httpsThreadfinDomain.description}}";
        break;
      case "bindIpAddress":
        text = "{{.settings.bindIpAddress.description}}";
        break;
      case "httpThreadfinDomain":
        text = "{{.settings.httpThreadfinDomain.description}}";
        break;
      case "enableNonAscii":
        text = "{{.settings.enableNonAscii.description}}";
        break;
      case "epgCategories":
        text = "{{.settings.epgCategories.description}}";
        break;
      case "epgCategoriesColors":
        text = "{{.settings.epgCategoriesColors.description}}";
        break;
      case "buffer.timeout":
        text = "{{.settings.bufferTimeout.description}}";
        break;
      case "user.agent":
        text = "{{.settings.userAgent.description}}";
        break;
      case "ffmpeg.path":
        text = "{{.settings.ffmpegPath.description}}";
        break;
      case "ffmpeg.options":
        text = "{{.settings.ffmpegOptions.description}}";
        break;
      case "ffmpeg.forceHttp":
        text = "{{.settings.ffmpegForceHttp.description}}";
        break;
      case "vlc.path":
        text = "{{.settings.vlcPath.description}}";
        break;
      case "vlc.options":
        text = "{{.settings.vlcOptions.description}}";
        break;
      case "epgSource":
        text = "{{.settings.epgSource.description}}";
        break;
      case "tuner":
        text = "{{.settings.tuner.description}}";
        break;
      case "update":
        text = "{{.settings.update.description}}";
        break;
      case "api":
        text = "{{.settings.api.description}}";
        break;
      case "ssdp":
        text = "{{.settings.ssdp.description}}";
        break;
      case "files.update":
        text = "{{.settings.filesUpdate.description}}";
        break;
      case "cache.images":
        text = "{{.settings.cacheImages.description}}";
        break;
      case "xepg.replace.missing.images":
        text = "{{.settings.replaceEmptyImages.description}}";
        break;
      case "xepg.replace.channel.title":
        text = "{{.settings.replaceChannelTitle.description}}";
        break;
      case "udpxy":
        text = "{{.settings.udpxy.description}}";
        break;
      default:
        text = "";
        break;
    }
    var tdLeft = document.createElement("TD");
    tdLeft.innerHTML = "";
    var tdRight = document.createElement("TD");
    var pre = document.createElement("PRE");
    pre.innerHTML = text;
    tdRight.appendChild(pre);
    description.appendChild(tdLeft);
    description.appendChild(tdRight);
    return description;
  }
};
var SettingsCategoryItem = class extends SettingsCategory {
  constructor(headline, settingsKeys) {
    super();
    this.headline = headline;
    this.settingsKeys = settingsKeys;
  }
  createCategory() {
    var headline = this.createCategoryHeadline(this.headline);
    var settingsKeys = this.settingsKeys;
    var doc = document.getElementById(this.DocumentID);
    doc.appendChild(headline);
    var table = document.createElement("TABLE");
    var keys = settingsKeys.split(",");
    keys.forEach((settingsKey) => {
      switch (settingsKey) {
        case "authentication.pms":
        case "authentication.m3u":
        case "authentication.xml":
        case "authentication.api":
          if (SERVER["settings"]["authentication.web"] == false) {
            break;
          }
        default:
          var item = this.createSettings(settingsKey);
          var description = this.createDescription(settingsKey);
          table.appendChild(item);
          table.appendChild(description);
          break;
      }
    });
    doc.appendChild(table);
    doc.appendChild(this.createHR());
  }
};
function saveSettings() {
  var cmd = "saveSettings";
  var div = document.getElementById("content_settings");
  var controls = Array.prototype.slice.call(div.querySelectorAll("input, select"));
  var newSettings = serializeSettingsChanges(controls);
  var data = new Object();
  data["settings"] = newSettings;
  var server = new Server(cmd);
  server.request(data);
}
var SETTINGS_SECTIONS = [
  {
    key: "general",
    label: "General",
    settings: ["epgSource", "ThreadfinAutoUpdate", "ssdp", "tuner", "epgCategories", "epgCategoriesColors", "dummy", "dummyChannel", "ignoreFilters", "api"]
  },
  {
    key: "files",
    label: "Files",
    settings: ["update", "files.update", "temp.path", "cache.images", "bindIpAddress", "httpThreadfinDomain", "forceHttps", "excludeStreamHttps", "httpsPort", "httpsThreadfinDomain", "xepg.replace.missing.images", "xepg.replace.channel.title", "enableNonAscii"]
  },
  {
    key: "streaming",
    label: "Streaming",
    settings: ["udpxy", "buffer.size.kb", "buffer.timeout", "user.agent", "ffmpeg.path", "ffmpeg.options", "ffmpeg.forceHttp", "vlc.path", "vlc.options"]
  },
  { key: "backup", label: "Backup", settings: ["backup.path", "backup.keep"] },
  {
    key: "authentication",
    label: "Authentication",
    settings: ["authentication.web", "authentication.pms", "authentication.m3u", "authentication.xml", "authentication.api"]
  }
];
var SETTINGS_NUMERIC_KEYS = {
  "tuner": true,
  "backup.keep": true,
  "buffer.size.kb": true,
  "buffer.timeout": true,
  "httpsPort": true
};
function serializeSettingsChanges(controls) {
  var changes = {};
  controls.forEach((control) => {
    var classes = String(control.className || "").split(/\s+/);
    if (classes.indexOf("changed") == -1 || !control.name) {
      return;
    }
    var name = String(control.name);
    if (String(control.tagName).toUpperCase() == "INPUT" && control.type == "checkbox") {
      changes[name] = Boolean(control.checked);
      return;
    }
    var value = String(control.value == void 0 ? "" : control.value);
    if (name == "update") {
      changes[name] = value.split(",").filter((entry) => entry.length > 0);
      return;
    }
    if (SETTINGS_NUMERIC_KEYS[name]) {
      changes[name] = name == "buffer.timeout" ? parseFloat(value) : parseInt(value, 10);
      return;
    }
    changes[name] = value;
  });
  return changes;
}
function settingsEPGImpact(value) {
  if (value == "XEPG") {
    return "XEPG uses XMLTV guide sources and makes Mapping plus M3U and XMLTV outputs applicable after you save.";
  }
  return "PMS lets the connected client manage guide data; XMLTV sources, Mapping, and M3U/XMLTV guide outputs are not used by this mode.";
}
function renderEPGSourcePreview(value, target) {
  if (target) {
    target.textContent = settingsEPGImpact(value);
  }
}
function activateSettingsSection(sectionKey, focusActiveTab) {
  var root = document.querySelector(".tf-settings");
  if (!root) {
    return;
  }
  var panels = root.querySelectorAll("[data-settings-panel]");
  for (var i = 0; i < panels.length; i++) {
    var panel = panels[i];
    panel.hidden = panel.getAttribute("data-settings-panel") != sectionKey;
  }
  var tabs = root.querySelectorAll("[data-settings-tab]");
  for (var j = 0; j < tabs.length; j++) {
    var tab = tabs[j];
    var active = tab.getAttribute("data-settings-tab") == sectionKey;
    tab.setAttribute("aria-selected", active ? "true" : "false");
    tab.setAttribute("tabindex", active ? "0" : "-1");
    if (active && focusActiveTab) {
      tab.focus();
    }
  }
}
function handleSettingsTabKeydown(event, currentIndex) {
  var nextIndex = currentIndex;
  switch (event.key) {
    case "ArrowRight":
    case "ArrowDown":
      nextIndex = (currentIndex + 1) % SETTINGS_SECTIONS.length;
      break;
    case "ArrowLeft":
    case "ArrowUp":
      nextIndex = (currentIndex + SETTINGS_SECTIONS.length - 1) % SETTINGS_SECTIONS.length;
      break;
    case "Home":
      nextIndex = 0;
      break;
    case "End":
      nextIndex = SETTINGS_SECTIONS.length - 1;
      break;
    default:
      return;
  }
  event.preventDefault();
  activateSettingsSection(SETTINGS_SECTIONS[nextIndex].key, true);
}
function syncSettingsAuthenticationGating(root) {
  var web = root.querySelector('[name="authentication.web"]');
  var explanation = root.querySelector(".tf-settings-auth-gating");
  var enabled = web ? web.checked : false;
  if (explanation) {
    explanation.textContent = enabled ? "WEB access is enabled, so PMS, M3U, XML, and API permissions can be configured." : "WEB access gates user management. Saving WEB as disabled clears PMS, M3U, XML, and API authentication on the server.";
  }
  ;
  ["authentication.pms", "authentication.m3u", "authentication.xml", "authentication.api"].forEach((key) => {
    var input = root.querySelector('[name="' + key + '"]');
    if (input) {
      input.disabled = !enabled;
    }
  });
}
function labelSettingsControls(root) {
  var controls = root.querySelectorAll("input, select, textarea");
  for (var index = 0; index < controls.length; index++) {
    var control = controls[index];
    if (String(control.type || "").toLowerCase() == "hidden") {
      continue;
    }
    var row = control.closest("tr");
    var labelCell = row ? row.querySelector("td:first-child") : null;
    if (!labelCell || !labelCell.textContent) {
      continue;
    }
    var key = String(control.name || "setting").replace(/[^A-Za-z0-9_-]+/g, "-");
    labelCell.id = "settings-field-label-" + key + "-" + index;
    control.id = "settings-field-" + key + "-" + index;
    control.setAttribute("aria-labelledby", labelCell.id);
  }
}
function renderSettingsPage(doc) {
  doc.innerHTML = "";
  var root = document.createElement("div");
  root.className = "tf-settings";
  var header = document.createElement("header");
  header.className = "tf-admin-header";
  var headingGroup = document.createElement("div");
  var heading = document.createElement("h1");
  heading.textContent = "Settings";
  var purpose = document.createElement("p");
  purpose.textContent = "Configure Threadfin ingestion, files, streaming, backups, and access without changing unrelated settings.";
  headingGroup.appendChild(heading);
  headingGroup.appendChild(purpose);
  var save = document.createElement("button");
  save.type = "button";
  save.className = "tf-primary-action";
  save.textContent = "Save settings";
  save.addEventListener("click", saveSettings);
  header.appendChild(headingGroup);
  header.appendChild(save);
  root.appendChild(header);
  var nav = document.createElement("div");
  nav.className = "tf-settings-nav";
  nav.setAttribute("role", "tablist");
  nav.setAttribute("aria-label", "Settings sections");
  SETTINGS_SECTIONS.forEach((definition, index) => {
    var button = document.createElement("button");
    button.type = "button";
    button.textContent = definition.label;
    button.setAttribute("role", "tab");
    button.id = "settings-tab-" + definition.key;
    button.setAttribute("data-settings-tab", definition.key);
    button.setAttribute("aria-controls", "settings-panel-" + definition.key);
    button.setAttribute("aria-selected", index == 0 ? "true" : "false");
    button.setAttribute("tabindex", index == 0 ? "0" : "-1");
    button.addEventListener("click", () => activateSettingsSection(definition.key));
    button.addEventListener("keydown", (event) => handleSettingsTabKeydown(event, index));
    nav.appendChild(button);
  });
  root.appendChild(nav);
  var settingsHost = document.createElement("div");
  settingsHost.id = "content_settings";
  settingsHost.className = "tf-settings-panels";
  SETTINGS_SECTIONS.forEach((definition, index) => {
    var panel = document.createElement("section");
    panel.id = "settings-panel-" + definition.key;
    panel.className = "tf-settings-panel";
    panel.setAttribute("role", "tabpanel");
    panel.setAttribute("aria-labelledby", "settings-tab-" + definition.key);
    panel.setAttribute("data-settings-panel", definition.key);
    panel.hidden = index != 0;
    var panelHeading = document.createElement("h2");
    panelHeading.textContent = definition.label;
    panel.appendChild(panelHeading);
    if (definition.key == "authentication") {
      var gating = document.createElement("p");
      gating.className = "tf-settings-auth-gating";
      panel.appendChild(gating);
    }
    var table = document.createElement("table");
    table.className = "tf-settings-table";
    var renderer = new SettingsCategory();
    definition.settings.forEach((settingsKey) => {
      var setting = renderer.createSettings(settingsKey);
      setting.setAttribute("data-setting-key", settingsKey);
      table.appendChild(setting);
      table.appendChild(renderer.createDescription(settingsKey));
    });
    panel.appendChild(table);
    if (definition.key == "general") {
      var preview = document.createElement("p");
      preview.className = "tf-settings-epg-preview";
      preview.setAttribute("aria-live", "polite");
      panel.insertBefore(preview, table.nextSibling);
      var epgSource = table.querySelector('[name="epgSource"]');
      if (epgSource) {
        renderEPGSourcePreview(epgSource.value, preview);
        epgSource.addEventListener("change", () => renderEPGSourcePreview(epgSource.value, preview));
      }
    }
    if (definition.key == "backup") {
      var backupActions = document.createElement("div");
      backupActions.className = "tf-backup-actions";
      var backupButton = document.createElement("button");
      backupButton.type = "button";
      backupButton.textContent = "Download backup";
      backupButton.addEventListener("click", backup);
      var restoreButton = document.createElement("button");
      restoreButton.type = "button";
      restoreButton.className = "tf-destructive-action";
      restoreButton.textContent = "Restore backup\u2026";
      restoreButton.addEventListener("click", restore);
      backupActions.appendChild(backupButton);
      backupActions.appendChild(restoreButton);
      panel.appendChild(backupActions);
    }
    settingsHost.appendChild(panel);
  });
  root.appendChild(settingsHost);
  doc.appendChild(root);
  labelSettingsControls(root);
  var authWeb = root.querySelector('[name="authentication.web"]');
  if (authWeb) {
    authWeb.addEventListener("change", () => syncSettingsAuthenticationGating(root));
  }
  syncSettingsAuthenticationGating(root);
}
function showLogs(bottom) {
  var logs = SERVER["log"] && Array.isArray(SERVER["log"]["log"]) ? SERVER["log"]["log"] : [];
  var div = document.getElementById("content_log_entries");
  if (!div) {
    return;
  }
  renderLogEntryList(div, logs);
  setTimeout(function() {
    if (bottom == true) {
      var wrapper = document.getElementById("box-wrapper");
      wrapper.scrollTop = wrapper.scrollHeight;
    }
  }, 10);
}
function resetLogs() {
  if (!confirm("Reset all current log entries? This permanently removes them and cannot be undone.")) {
    return;
  }
  var cmd = "resetLogs";
  var data = new Object();
  var server = new Server(cmd);
  server.request(data);
}
var LOG_VIEW_STATE = {
  search: "",
  severities: { DEBUG: false, WARNING: false, ERROR: false }
};
function logTextForDisplay(entry) {
  return String(entry == void 0 ? "" : entry).replace(/&nbsp;/g, " ");
}
function logEntrySeverity(entry) {
  var text = String(entry || "");
  if (text.indexOf("[ERROR]") != -1) {
    return "ERROR";
  }
  if (text.indexOf("[WARNING]") != -1) {
    return "WARNING";
  }
  if (text.indexOf("[DEBUG]") != -1) {
    return "DEBUG";
  }
  return "";
}
function logEntryLayer(entry) {
  var match = String(entry || "").match(/(?:\[Threadfin\]|\[DEBUG\])\s+(?:\[(?:WARNING|ERROR)\]\s+)?([A-Za-z][A-Za-z0-9 /_-]{0,39}):(?:&nbsp;|\s)/);
  return match ? match[1].trim() : "";
}
function logEntryMatches(entry) {
  var search = LOG_VIEW_STATE.search.trim().toLocaleLowerCase();
  if (search && logTextForDisplay(entry).toLocaleLowerCase().indexOf(search) == -1) {
    return false;
  }
  var active = Object.keys(LOG_VIEW_STATE.severities).filter((marker) => LOG_VIEW_STATE.severities[marker]);
  return active.length == 0 || active.indexOf(logEntrySeverity(entry)) != -1;
}
function createLogEntryElement(entry) {
  var text = logTextForDisplay(entry);
  var element = document.createElement("article");
  element.className = "tf-log-entry";
  var severity = logEntrySeverity(entry);
  if (severity) {
    element.setAttribute("data-severity", severity);
  }
  var layer = logEntryLayer(entry);
  if (layer) {
    var layerLabel = document.createElement("span");
    layerLabel.className = "tf-log-layer";
    layerLabel.textContent = layer;
    element.appendChild(layerLabel);
  }
  var pre = document.createElement("pre");
  pre.textContent = text;
  element.appendChild(pre);
  element.queryText = text;
  return element;
}
function renderLogEntryList(host, entries) {
  var visible = (entries || []).filter(logEntryMatches).map(createLogEntryElement);
  host.replaceChildren.apply(host, visible);
  if (visible.length == 0) {
    var empty = document.createElement("p");
    empty.className = "tf-log-empty";
    empty.textContent = "No log entries match the current filters.";
    host.appendChild(empty);
  }
}
function renderLogPage(doc) {
  doc.innerHTML = "";
  var root = document.createElement("div");
  root.id = "content_log";
  root.className = "tf-log";
  var header = document.createElement("header");
  header.className = "tf-admin-header";
  var headingGroup = document.createElement("div");
  var heading = document.createElement("h1");
  heading.textContent = "Log";
  var purpose = document.createElement("p");
  purpose.textContent = "Search Threadfin's current diagnostic record or limit it to response-backed DEBUG, WARNING, and ERROR markers.";
  headingGroup.appendChild(heading);
  headingGroup.appendChild(purpose);
  header.appendChild(headingGroup);
  root.appendChild(header);
  var controls = document.createElement("div");
  controls.className = "tf-log-controls";
  var searchLabel = document.createElement("label");
  searchLabel.textContent = "Search log";
  var search = document.createElement("input");
  search.type = "search";
  search.id = "log-search";
  search.value = LOG_VIEW_STATE.search;
  search.autocomplete = "off";
  search.addEventListener("input", () => {
    LOG_VIEW_STATE.search = search.value;
    showLogs(false);
  });
  searchLabel.appendChild(search);
  controls.appendChild(searchLabel);
  var severityGroup = document.createElement("fieldset");
  var legend = document.createElement("legend");
  legend.textContent = "Severity markers";
  severityGroup.appendChild(legend);
  ["DEBUG", "WARNING", "ERROR"].forEach((marker) => {
    var label = document.createElement("label");
    var input = document.createElement("input");
    input.type = "checkbox";
    input.checked = LOG_VIEW_STATE.severities[marker];
    input.addEventListener("change", () => {
      LOG_VIEW_STATE.severities[marker] = input.checked;
      showLogs(false);
    });
    label.appendChild(input);
    label.appendChild(document.createTextNode(marker));
    severityGroup.appendChild(label);
  });
  controls.appendChild(severityGroup);
  root.appendChild(controls);
  var wrapper = document.createElement("div");
  wrapper.id = "box-wrapper";
  wrapper.className = "tf-log-scroll";
  var entries = document.createElement("div");
  entries.id = "content_log_entries";
  entries.className = "tf-log-entries";
  wrapper.appendChild(entries);
  root.appendChild(wrapper);
  var danger = document.createElement("section");
  danger.className = "tf-log-danger-zone";
  var dangerHeading = document.createElement("h2");
  dangerHeading.textContent = "Reset log";
  var dangerText = document.createElement("p");
  dangerText.textContent = "Permanently remove all current in-memory log entries. Filtering does not remove entries.";
  var reset = document.createElement("button");
  reset.type = "button";
  reset.className = "tf-destructive-action";
  reset.textContent = "Reset logs\u2026";
  reset.addEventListener("click", resetLogs);
  danger.appendChild(dangerHeading);
  danger.appendChild(dangerText);
  danger.appendChild(reset);
  root.appendChild(danger);
  doc.appendChild(root);
  showLogs(true);
}
var SERVER = new Object();
var BULK_EDIT = false;
var COLUMN_TO_SORT;
var INACTIVE_COLUMN_TO_SORT;
var SEARCH_MAPPING = new Object();
var UNDO = new Object();
var WS_AVAILABLE = false;
var tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]');
var tooltipList = [...tooltipTriggerList].map((tooltipTriggerEl) => new Tooltip(tooltipTriggerEl));
var clipboard = new import_clipboard.default(".copy-btn");
clipboard.on("success", function(e) {
  const tooltip = Tooltip.getInstance(e.trigger);
  if (tooltip)
    tooltip.setContent({ ".tooltip-inner": "Copied!" });
});
clipboard.on("error", function(e) {
  console.log(e);
});
var popupModal = new Dialog(document.getElementById("popup"));
var menuItems = new Array();
menuItems.push(new MainMenuItem("playlist", "{{.mainMenu.item.playlist}}"));
menuItems.push(new MainMenuItem("xmltv", "{{.mainMenu.item.xmltv}}"));
menuItems.push(new MainMenuItem("filter", "{{.mainMenu.item.filter}}"));
menuItems.push(new MainMenuItem("mapping", "{{.mainMenu.item.mapping}}"));
menuItems.push(new MainMenuItem("users", "{{.mainMenu.item.users}}"));
menuItems.push(new MainMenuItem("settings", "{{.mainMenu.item.settings}}"));
menuItems.push(new MainMenuItem("log", "{{.mainMenu.item.log}}"));
menuItems.push(new MainMenuItem("logout", "{{.mainMenu.item.logout}}"));
var settingsCategory = new Array();
settingsCategory.push(new SettingsCategoryItem("{{.settings.category.general}}", "epgSource,ThreadfinAutoUpdate,ssdp,tuner,epgCategories,epgCategoriesColors,dummy,dummyChannel,ignoreFilters,api"));
settingsCategory.push(new SettingsCategoryItem("{{.settings.category.files}}", "update,files.update,temp.path,cache.images,bindIpAddress,httpThreadfinDomain,forceHttps,excludeStreamHttps,httpsPort,httpsThreadfinDomain,xepg.replace.missing.images,xepg.replace.channel.title,enableNonAscii"));
settingsCategory.push(new SettingsCategoryItem("{{.settings.category.streaming}}", "udpxy,buffer.size.kb,buffer.timeout,user.agent,ffmpeg.path,ffmpeg.options,ffmpeg.forceHttp,vlc.path,vlc.options"));
settingsCategory.push(new SettingsCategoryItem("{{.settings.category.backup}}", "backup.path,backup.keep"));
settingsCategory.push(new SettingsCategoryItem("{{.settings.category.authentication}}", "authentication.web,authentication.pms,authentication.m3u,authentication.xml,authentication.api"));
function showPopUpElement(elm) {
  showElement(elm, true);
  return;
}
function showElement(elmID, type) {
  if (elmID == "popup-custom" || elmID == "popup") {
    switch (type) {
      case true:
        popupModal.show();
        break;
      case false:
        popupModal.hide();
        break;
    }
  }
  if (elmID == "loading") {
    switch (type) {
      case true:
        document.getElementById("loading").hidden = false;
        break;
      case false:
        document.getElementById("loading").hidden = true;
        break;
    }
  }
}
function getLocalData(dataType, id) {
  var data = new Object();
  switch (dataType) {
    case "m3u":
      data = SERVER["settings"]["files"][dataType][id];
      break;
    case "hdhr":
      data = SERVER["settings"]["files"][dataType][id];
      break;
    case "filter":
    case "custom-filter":
    case "group-title":
      if (id == -1) {
        data["active"] = true;
        data["liveEvent"] = false;
        data["caseSensitive"] = false;
        data["description"] = "";
        data["exclude"] = "";
        data["filter"] = "";
        data["include"] = "";
        data["name"] = "";
        data["type"] = "group-title";
        data["x-category"] = "";
        SERVER["settings"]["filter"][id] = data;
      }
      data = SERVER["settings"]["filter"][id];
      break;
    case "xmltv":
      data = SERVER["settings"]["files"][dataType][id];
      break;
    case "users":
      data = SERVER["users"][id]["data"];
      break;
    case "mapping":
      data = SERVER["xepg"]["epgMapping"][id];
      break;
    case "m3uGroups":
      data = SERVER["data"]["playlist"]["m3u"]["groups"];
      break;
  }
  return data;
}
function getObjKeys(obj) {
  var keys = new Array();
  for (var i in obj) {
    if (obj.hasOwnProperty(i)) {
      keys.push(i);
    }
  }
  return keys;
}
function getOwnObjProps(object) {
  return object ? Object.getOwnPropertyNames(object) : [];
}
function getAllSelectedChannels() {
  var channels = new Array();
  if (BULK_EDIT == false) {
    return channels;
  }
  var trs = document.getElementById("content_table").getElementsByTagName("TR");
  for (var i = 1; i < trs.length; i++) {
    if (trs[i].style.display != "none") {
      if (trs[i].firstChild.firstChild.checked == true) {
        channels.push(trs[i].id);
      }
    }
  }
  var trs_inactive = document.getElementById("inactive_content_table").getElementsByTagName("TR");
  for (var i = 1; i < trs_inactive.length; i++) {
    if (trs_inactive[i].style.display != "none") {
      if (trs_inactive[i].firstChild.firstChild.checked == true) {
        channels.push(trs_inactive[i].id);
      }
    }
  }
  return channels;
}
function sortTable(column, table_name = "content_table") {
  if (column == COLUMN_TO_SORT && table_name == "content_table" || column == INACTIVE_COLUMN_TO_SORT && table_name == "inactive_content_table") {
    return;
  }
  var table = document.getElementById(table_name);
  var tableHead = table.getElementsByTagName("TR")[0];
  var tableItems = tableHead.getElementsByTagName("TD");
  var sortObj = new Object();
  var x, xValue;
  var tableHeader;
  var sortByString = false;
  if (column > 0 && COLUMN_TO_SORT > 0 && table_name == "content_table") {
    tableItems[COLUMN_TO_SORT].className = "pointer";
    tableItems[column].className = "sortThis";
  } else if (column > 0 && INACTIVE_COLUMN_TO_SORT > 0 && table_name == "inactive_content_table") {
    tableItems[INACTIVE_COLUMN_TO_SORT].className = "pointer";
    tableItems[column].className = "sortThis";
  }
  if (table_name == "content_table") {
    COLUMN_TO_SORT = column;
  } else if (table_name == "inactive_content_table") {
    INACTIVE_COLUMN_TO_SORT = column;
  }
  var rows = table.rows;
  if (rows[1] != void 0) {
    tableHeader = rows[0];
    x = rows[1].getElementsByTagName("TD")[column];
    for (i = 1; i < rows.length; i++) {
      x = rows[i].getElementsByTagName("TD")[column];
      switch (x.childNodes[0].tagName.toLowerCase()) {
        case "input":
          xValue = x.getElementsByTagName("INPUT")[0].value.toLowerCase();
          break;
        case "p":
          xValue = x.getElementsByTagName("P")[0].innerText.toLowerCase();
          break;
        default:
          console.log(x.childNodes[0].tagName);
      }
      if (xValue == "") {
        xValue = i;
        sortObj[i] = rows[i];
      } else {
        switch (isNaN(xValue)) {
          case false:
            xValue = parseFloat(xValue);
            sortObj[xValue] = rows[i];
            break;
          case true:
            sortByString = true;
            sortObj[xValue.toLowerCase() + i] = rows[i];
            break;
        }
      }
    }
    while (table.firstChild) {
      table.removeChild(table.firstChild);
    }
    var sortValues = getObjKeys(sortObj);
    if (sortByString == true) {
      if (column == 3) {
        var collator = new Intl.Collator(void 0, { numeric: true, sensitivity: "base" });
        sortValues.sort(collator.compare);
      } else {
        sortValues.sort();
      }
    } else {
      let sortFloat = function(a, b) {
        return a - b;
      };
      sortValues.sort(sortFloat);
    }
    table.appendChild(tableHeader);
    for (var i = 0; i < sortValues.length; i++) {
      table.appendChild(sortObj[sortValues[i]]);
    }
  }
  return;
}
function createSearchObj() {
  SEARCH_MAPPING = new Object();
  var data = SERVER["xepg"]["epgMapping"];
  var channels = getObjKeys(data);
  var channelKeys = ["x-active", "x-channelID", "x-name", "_file.m3u.name", "x-group-title", "x-xmltv-file"];
  channels.forEach((id) => {
    channelKeys.forEach((key) => {
      if (key == "x-active") {
        switch (data[id][key]) {
          case true:
            SEARCH_MAPPING[id] = "online ";
            break;
          case false:
            SEARCH_MAPPING[id] = "offline ";
            break;
        }
      } else {
        if (key == "x-xmltv-file") {
          var xmltvFile = getValueFromProviderFile(data[id][key], "xmltv", "name");
          if (xmltvFile != void 0) {
            SEARCH_MAPPING[id] = SEARCH_MAPPING[id] + xmltvFile + " ";
          }
        } else {
          SEARCH_MAPPING[id] = SEARCH_MAPPING[id] + data[id][key] + " ";
        }
      }
    });
  });
  return;
}
function searchInMapping() {
  var searchValue = document.getElementById("searchMapping").value;
  var trs = document.getElementById("content_table").getElementsByTagName("TR");
  for (var i = 1; i < trs.length; ++i) {
    var id = trs[i].getAttribute("id");
    var element = SEARCH_MAPPING[id];
    switch (element.toLowerCase().includes(searchValue.toLowerCase())) {
      case true:
        document.getElementById(id).style.display = "";
        break;
      case false:
        document.getElementById(id).style.display = "none";
        break;
    }
  }
  return;
}
function changeChannelNumbers(elements) {
  var starting_number_element = document.getElementsByName("x-channels-start")[0];
  var elems = elements.split(",");
  var starting_number = parseFloat(starting_number_element.value);
  var data = SERVER["xepg"]["epgMapping"];
  elems.forEach((element) => {
    var elem = document.getElementById(element);
    var input = elem.childNodes[1].firstChild;
    input.value = starting_number.toString();
    data[element]["x-channelID"] = starting_number.toString();
    starting_number++;
  });
  if (COLUMN_TO_SORT == 1) {
    COLUMN_TO_SORT = -1;
    sortTable(1);
  }
  if (INACTIVE_COLUMN_TO_SORT == 1) {
    INACTIVE_COLUMN_TO_SORT = -1;
    sortTable(1, "inactive_content_page");
  }
}
function backup() {
  var data = new Object();
  var cmd = "ThreadfinBackup";
  var server = new Server(cmd);
  server.request(data);
  return;
}
function toggleChannelStatus(id) {
  var element;
  var status;
  if (document.getElementById("active")) {
    var checkbox = document.getElementById("active");
    status = checkbox.checked;
  }
  var ids = getAllSelectedChannels();
  if (ids.length == 0) {
    ids.push(id);
  }
  ids.forEach((id2) => {
    var channel = SERVER["xepg"]["epgMapping"][id2];
    channel["x-active"] = status;
    switch (channel["x-active"]) {
      case true:
        if (channel["x-xmltv-file"] == "-" || channel["x-mapping"] == "-") {
          if (BULK_EDIT == false) {
            checkbox.checked = true;
          }
          channel["x-active"] = true;
        }
        break;
      case false:
        break;
    }
    if (channel["x-active"] == false) {
      document.getElementById(id2).className = "notActiveEPG";
    } else {
      document.getElementById(id2).className = "activeEPG";
    }
  });
}
function restore() {
  if (document.getElementById("upload")) {
    document.getElementById("upload").remove();
  }
  var restore2 = document.createElement("INPUT");
  restore2.setAttribute("type", "file");
  restore2.setAttribute("class", "notVisible");
  restore2.setAttribute("name", "");
  restore2.id = "upload";
  document.body.appendChild(restore2);
  restore2.click();
  restore2.onchange = function() {
    var filename = restore2.files[0].name;
    var check = confirm("Restore " + filename + "?\nAll current Threadfin data will be replaced with this backup. This action cannot be undone.");
    if (check == true) {
      var reader = new FileReader();
      var file = document.querySelector("input[type=file]").files[0];
      if (file) {
        reader.readAsDataURL(file);
        reader.onload = function() {
          var data = new Object();
          var cmd = "ThreadfinRestore";
          data["base64"] = reader.result;
          var server = new Server(cmd);
          server.request(data);
        };
      } else {
        alert("File could not be loaded");
      }
      restore2.remove();
      return;
    }
  };
  return;
}
function uploadLogo() {
  if (document.getElementById("upload")) {
    document.getElementById("upload").remove();
  }
  var upload = document.createElement("INPUT");
  upload.setAttribute("type", "file");
  upload.setAttribute("class", "notVisible");
  upload.setAttribute("name", "");
  upload.id = "upload";
  document.body.appendChild(upload);
  upload.click();
  upload.onblur = function() {
    alert();
  };
  upload.onchange = function() {
    var filename = upload.files[0].name;
    var reader = new FileReader();
    var file = document.querySelector("input[type=file]").files[0];
    if (file) {
      reader.readAsDataURL(file);
      reader.onload = function() {
        var data = new Object();
        var cmd = "uploadLogo";
        data["base64"] = reader.result;
        data["filename"] = file.name;
        var server = new Server(cmd);
        server.request(data);
        var updateLogo = document.getElementById("update-icon");
        updateLogo.checked = false;
        updateLogo.className = "changed";
      };
    } else {
      alert("File could not be loaded");
    }
    upload.remove();
    return;
  };
}
function probeChannel(url) {
  if (document.getElementById("probeDetails")) {
    document.getElementById("probeDetails").innerHTML = "Probing Channel Details...";
  }
  var data = new Object();
  var cmd = "probeChannel";
  data["probeUrl"] = url;
  var server = new Server(cmd);
  server.request(data);
  return;
}
function updateLog() {
  console.log("TOKEN");
  var server = new Server("updateLog");
  server.request(new Object());
}
PageReady();
/*! Bundled license information:

clipboard/dist/clipboard.js:
  (*!
   * clipboard.js v2.0.11
   * https://clipboardjs.com/
   *
   * Licensed MIT © Zeno Rocha
   *)

bootstrap/js/dist/dom/data.js:
  (*!
  * Bootstrap data.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/dom/event-handler.js:
  (*!
  * Bootstrap event-handler.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/dom/manipulator.js:
  (*!
  * Bootstrap manipulator.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/util/index.js:
  (*!
  * Bootstrap index.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/util/config.js:
  (*!
  * Bootstrap config.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/base-component.js:
  (*!
  * Bootstrap base-component.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/dom/selector-engine.js:
  (*!
  * Bootstrap selector-engine.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/dialog-base.js:
  (*!
  * Bootstrap dialog-base.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/util/component-functions.js:
  (*!
  * Bootstrap component-functions.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/dialog.js:
  (*!
  * Bootstrap dialog.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/util/swipe.js:
  (*!
  * Bootstrap swipe.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/drawer.js:
  (*!
  * Bootstrap drawer.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/util/floating-ui.js:
  (*!
  * Bootstrap floating-ui.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/menu.js:
  (*!
  * Bootstrap menu.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/combobox.js:
  (*!
  * Bootstrap combobox.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/util/sanitizer.js:
  (*!
  * Bootstrap sanitizer.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/chips.js:
  (*!
  * Bootstrap chips.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/util/template-factory.js:
  (*!
  * Bootstrap template-factory.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)

bootstrap/js/dist/tooltip.js:
  (*!
  * Bootstrap tooltip.js v6.0.0-alpha.1 (https://getbootstrap.com/)
  * Copyright 2011-2026 The Bootstrap Authors (https://github.com/twbs/bootstrap/graphs/contributors)
  * Licensed under MIT (https://github.com/twbs/bootstrap/blob/main/LICENSE)
  *)
*/
