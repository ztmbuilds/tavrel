declare module "*.svg?component" {
  const component: import("vue").FunctionalComponent<
    import("vue").SVGAttributes
  >;
  export default component;
}
