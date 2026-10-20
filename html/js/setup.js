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
var WizardCategory = class {
  constructor() {
    this.DocumentID = "content";
  }
  createCategoryHeadline(value) {
    var element = document.createElement("H2");
    element.textContent = value;
    return element;
  }
};
var WizardItem = class extends WizardCategory {
  constructor(key, headline) {
    super();
    this.headline = headline;
    this.key = key;
  }
  createWizard() {
    var headline = this.createCategoryHeadline(this.headline);
    headline.id = "wizard-field-label";
    var key = this.key;
    var content = new PopupContent();
    var description;
    var wizardField;
    var doc = document.getElementById(this.DocumentID);
    doc.innerHTML = "";
    doc.appendChild(headline);
    switch (key) {
      case "tuner":
        var text = new Array();
        var values = new Array();
        for (var i = 1; i <= 100; i++) {
          text.push(i);
          values.push(i);
        }
        var select = content.createSelect(text, values, "1", key);
        select.setAttribute("class", "wizard");
        select.id = key;
        wizardField = select;
        doc.appendChild(select);
        description = "{{.wizard.tuner.description}}";
        break;
      case "epgSource":
        var text = ["PMS", "XEPG"];
        var values = ["PMS", "XEPG"];
        var select = content.createSelect(text, values, "XEPG", key);
        select.setAttribute("class", "wizard");
        select.id = key;
        wizardField = select;
        doc.appendChild(select);
        description = "{{.wizard.epgSource.description}}";
        break;
      case "m3u":
        var input = content.createInput("text", key, "");
        input.setAttribute("placeholder", "{{.wizard.m3u.placeholder}}");
        input.setAttribute("class", "wizard");
        input.id = key;
        wizardField = input;
        doc.appendChild(input);
        description = "{{.wizard.m3u.description}}";
        break;
      case "xmltv":
        var input = content.createInput("text", key, "");
        input.setAttribute("placeholder", "{{.wizard.xmltv.placeholder}}");
        input.setAttribute("class", "wizard");
        input.id = key;
        wizardField = input;
        doc.appendChild(input);
        description = "{{.wizard.xmltv.description}}";
        break;
      default:
        break;
    }
    if (wizardField) {
      wizardField.setAttribute("aria-labelledby", headline.id);
      wizardField.setAttribute("aria-describedby", "wizard-description wizard-field-error");
    }
    var help = document.createElement("div");
    help.id = "wizard-description";
    help.className = "tf-wizard-description";
    help.innerHTML = description;
    doc.appendChild(help);
    var fieldError = document.createElement("p");
    fieldError.id = "wizard-field-error";
    fieldError.className = "tf-wizard-field-error";
    fieldError.hidden = true;
    doc.appendChild(fieldError);
  }
};
function readyForConfiguration(wizard) {
  var server = new Server("getServerConfig");
  server.request(new Object());
  showElement("loading", false);
  showConfigurationWizard(wizard);
}
function showConfigurationWizard(wizard) {
  configurationWizard[wizard].createWizard();
  var visibleSteps = wizard == 1 ? configurationWizard.length : configurationWizardVisibleSteps();
  var progress = document.querySelectorAll("#wizard-progress li");
  Array.prototype.forEach.call(progress, function(item, index) {
    item.hidden = index >= visibleSteps;
    if (index == wizard) {
      item.setAttribute("aria-current", "step");
    } else {
      item.removeAttribute("aria-current");
    }
  });
  var step = document.getElementById("wizard-step-status");
  if (step) {
    step.textContent = "{{.wizard.progress}}".replace("{current}", String(wizard + 1)).replace("{total}", String(visibleSteps));
  }
  var next = document.getElementById("next");
  if (next) {
    next.disabled = false;
    next.value = wizard == visibleSteps - 1 ? "{{.wizard.finish}}" : "{{.button.next}}";
  }
  var requestStatus = document.getElementById("wizard-request-status");
  if (requestStatus) {
    requestStatus.textContent = "";
  }
}
function configurationWizardVisibleSteps() {
  var server = typeof SERVER == "object" && SERVER ? SERVER : {};
  var settings = server.settings && typeof server.settings == "object" ? server.settings : {};
  var clientInfo = server.clientInfo && typeof server.clientInfo == "object" ? server.clientInfo : {};
  var epgSource = String(settings.epgSource || clientInfo.epgSource || "XEPG").toUpperCase();
  return epgSource == "PMS" ? 3 : configurationWizard.length;
}
function saveWizard() {
  var cmd = "saveWizard";
  var div = document.getElementById("content");
  var config = div.getElementsByClassName("wizard");
  var wizard = new Object();
  var error = document.getElementById("wizard-field-error");
  if (error) {
    error.textContent = "";
    error.hidden = true;
  }
  var invalid = div.querySelectorAll('[aria-invalid="true"]');
  Array.prototype.forEach.call(invalid, function(field) {
    field.removeAttribute("aria-invalid");
  });
  for (var i = 0; i < config.length; i++) {
    var name;
    var value;
    switch (config[i].tagName) {
      case "SELECT":
        name = config[i].name;
        value = config[i].value;
        if (isNaN(value)) {
          wizard[name] = value;
        } else {
          wizard[name] = parseInt(value);
        }
        break;
      case "INPUT":
        switch (config[i].type) {
          case "text":
            name = config[i].name;
            value = config[i].value;
            if (value.length == 0) {
              showWizardFieldError(config[i], name.toUpperCase() + ": {{.alert.missingInput}}");
              return;
            }
            if ((name == "m3u" || name == "xmltv") && !sourceLocationAccepted(value)) {
              showWizardFieldError(config[i], "{{.sources.forms.locationInvalid}}");
              return;
            }
            wizard[name] = value;
            break;
        }
        break;
      default:
        break;
    }
  }
  var data = new Object();
  data["wizard"] = wizard;
  var requestStatus = document.getElementById("wizard-request-status");
  if (requestStatus) {
    requestStatus.textContent = "{{.wizard.saving}}";
  }
  var next = document.getElementById("next");
  if (next) {
    next.disabled = true;
  }
  var server = new Server(cmd);
  server.request(data);
}
function showWizardFieldError(field, message) {
  var error = document.getElementById("wizard-field-error");
  if (error) {
    error.textContent = message;
    error.hidden = false;
  }
  field.setAttribute("aria-invalid", "true");
  field.focus();
}
function completeConfigurationWizardRequest(response) {
  var status = document.getElementById("wizard-request-status");
  var next = document.getElementById("next");
  if (next) {
    next.disabled = false;
  }
  if (!status) {
    return;
  }
  status.textContent = response && response.status === true ? "{{.wizard.saved}}" : response && response.status === false ? sourceString(response.err) || "{{.sources.responseInvalid}}" : "{{.sources.responseInvalid}}";
}
function completeConfigurationWizard() {
  window.location.replace("/web/#overview");
  window.location.reload();
}
var configurationWizard = new Array();
configurationWizard.push(new WizardItem("tuner", "{{.wizard.tuner.title}}"));
configurationWizard.push(new WizardItem("epgSource", "{{.wizard.epgSource.title}}"));
configurationWizard.push(new WizardItem("m3u", "{{.wizard.m3u.title}}"));
configurationWizard.push(new WizardItem("xmltv", "{{.wizard.xmltv.title}}"));
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
var MainMenuItem = class {
  constructor(menuKey, value) {
    this.menuKey = menuKey;
    this.value = value;
  }
};
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
    var span = document.createElement("div");
    span.className = "tf-field-help";
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
var sourceFeedbackByKey = {};
var sourcePageFeedback = {};
function sourceRecord(value) {
  return value && typeof value == "object" && !Array.isArray(value) ? value : {};
}
function sourceString(value) {
  return typeof value == "string" ? value : value === void 0 || value === null ? "" : String(value);
}
function sourceFeedbackKey(providerType, id) {
  return providerType + ":" + id;
}
function sourceLocationAccepted(value) {
  var source = sourceString(value).trim();
  if (!source) {
    return false;
  }
  var scheme = source.match(/^([a-z][a-z0-9+.-]*):\/\//i);
  return !scheme || scheme[1].toLowerCase() == "http" || scheme[1].toLowerCase() == "https";
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
    var help = document.createElement("div");
    help.className = "tf-field-help";
    help.innerHTML = text;
    tdRight.appendChild(help);
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
var SERVER = new Object();
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
function getObjKeys(obj) {
  var keys = new Array();
  for (var i in obj) {
    if (obj.hasOwnProperty(i)) {
      keys.push(i);
    }
  }
  return keys;
}
document.getElementById("next").addEventListener("click", saveWizard);
readyForConfiguration(0);
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
