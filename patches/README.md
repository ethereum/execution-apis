# Documentation renderer patch

`@open-rpc+markdown-generator+0.2.0.patch` fixes the Node renderer used by the
Docusaurus plugin. It renders positional tuples and preserves their surrounding
array containers. This makes `trace_callMany` show its required `Calls` parameter,
the two tuple positions, and the length constraint on each tuple.

`npm ci` applies the patch through `postinstall`. If lifecycle scripts are disabled,
run `npm run postinstall` before generating or testing documentation. Patch failures
stop installation. The package version is locked; remove this patch when a renderer
release includes the fix, alongside the recursive/composed-schema tooling consolidation.

`npm run test:docs` checks every trace parameter's visibility and required status,
plus tuple labels, order and cardinality. The canonical OpenRPC schemas are unchanged
by the renderer patch.
