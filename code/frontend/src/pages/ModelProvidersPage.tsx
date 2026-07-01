import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { PageHeader } from "@/components/PageHeader"
import { api } from "@/api/client"
import type { ProviderType } from "@/types"

const PROVIDER_TYPE_LABEL: Record<ProviderType, string> = {
  openai_compatible: "OpenAI 兼容",
  anthropic: "Anthropic 原生",
}

export default function ModelProvidersPage() {
  const queryClient = useQueryClient()
  const providersQuery = useQuery({ queryKey: ["model-providers"], queryFn: api.listModelProviders })

  const [name, setName] = useState("")
  const [providerType, setProviderType] = useState<ProviderType>("openai_compatible")
  const [baseURL, setBaseURL] = useState("")
  const [apiKey, setApiKey] = useState("")
  const [modelName, setModelName] = useState("")

  const createMutation = useMutation({
    mutationFn: () =>
      api.createModelProvider({
        name,
        provider_type: providerType,
        base_url: baseURL || undefined,
        api_key: apiKey,
        model_name: modelName,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["model-providers"] })
      setName("")
      setBaseURL("")
      setApiKey("")
      setModelName("")
    },
  })

  return (
    <div>
      <PageHeader
        eyebrow="MODEL PROVIDERS"
        title="模型供应商"
        description="自己接入模型服务商的 API Key，构建 Agent 时选择用哪个供应商来跑对话。Key 只在创建时提交一次，落库前会用服务端密钥加密，列表和详情都不会再回显明文。"
      />

      <div className="mb-10 grid grid-cols-1 gap-6 lg:grid-cols-[1fr_360px]">
        <div>
          <h2 className="mb-4 font-display text-lg font-semibold">已配置的供应商</h2>
          {providersQuery.isLoading && <p className="text-sm text-muted-foreground">加载中…</p>}
          {providersQuery.error && (
            <Card className="border-destructive/40">
              <CardTitle className="text-destructive">加载失败</CardTitle>
              <CardDescription>{(providersQuery.error as Error).message}</CardDescription>
            </Card>
          )}
          {!providersQuery.isLoading && !providersQuery.error && (providersQuery.data?.length ?? 0) === 0 && (
            <Card>
              <CardTitle>还没有配置任何模型供应商</CardTitle>
              <CardDescription>先在右侧表单创建一个，构建 Agent 时才能选择使用哪个模型。</CardDescription>
            </Card>
          )}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {providersQuery.data?.map((p) => (
              <Card key={p.id}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <Badge variant="neutral">{PROVIDER_TYPE_LABEL[p.provider_type]}</Badge>
                    <span className={p.status === "enabled" ? "text-xs text-accent" : "text-xs text-muted-foreground"}>
                      {p.status === "enabled" ? "启用中" : "已禁用"}
                    </span>
                  </div>
                  <CardTitle>{p.name}</CardTitle>
                  <CardDescription>
                    模型：<span className="font-mono">{p.model_name}</span>
                    {p.base_url && (
                      <>
                        <br />
                        Base URL：<span className="font-mono">{p.base_url}</span>
                      </>
                    )}
                  </CardDescription>
                </CardHeader>
                <CardFooter>
                  <span>
                    ID：<span className="font-mono">{p.id}</span>
                  </span>
                </CardFooter>
              </Card>
            ))}
          </div>
        </div>

        <div className="glass-panel h-fit p-6">
          <p className="mb-4 text-sm font-medium">新建模型供应商</p>

          <label className="mb-1 block text-xs text-muted-foreground">名称</label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="例如：我的 Claude 账号"
            className="mb-3"
          />

          <label className="mb-1 block text-xs text-muted-foreground">协议类型</label>
          <div className="mb-3 flex gap-2">
            {(Object.keys(PROVIDER_TYPE_LABEL) as ProviderType[]).map((t) => (
              <button
                key={t}
                onClick={() => setProviderType(t)}
                className={`rounded-full border px-3 py-1.5 text-xs transition-colors ${
                  providerType === t
                    ? "border-primary/60 bg-primary/15 text-[#5aa6ff]"
                    : "border-glass-border text-muted-foreground"
                }`}
              >
                {PROVIDER_TYPE_LABEL[t]}
              </button>
            ))}
          </div>

          {providerType === "openai_compatible" && (
            <>
              <label className="mb-1 block text-xs text-muted-foreground">Base URL（必填）</label>
              <Input
                value={baseURL}
                onChange={(e) => setBaseURL(e.target.value)}
                placeholder="https://api.example.com/v1"
                className="mb-3"
              />
            </>
          )}
          {providerType === "anthropic" && (
            <>
              <label className="mb-1 block text-xs text-muted-foreground">Base URL（可选，留空用官方端点）</label>
              <Input
                value={baseURL}
                onChange={(e) => setBaseURL(e.target.value)}
                placeholder="留空默认使用 Anthropic 官方 API"
                className="mb-3"
              />
            </>
          )}

          <label className="mb-1 block text-xs text-muted-foreground">模型名</label>
          <Input
            value={modelName}
            onChange={(e) => setModelName(e.target.value)}
            placeholder={providerType === "anthropic" ? "例如：claude-sonnet-5" : "例如：gpt-4o-mini"}
            className="mb-3"
          />

          <label className="mb-1 block text-xs text-muted-foreground">API Key</label>
          <Input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder="sk-..."
            className="mb-4"
          />

          {createMutation.error && (
            <p className="mb-3 text-xs text-destructive">{(createMutation.error as Error).message}</p>
          )}

          <Button
            className="w-full"
            disabled={!name || !modelName || !apiKey || createMutation.isPending}
            onClick={() => createMutation.mutate()}
          >
            {createMutation.isPending ? "创建中…" : "创建供应商"}
          </Button>
        </div>
      </div>
    </div>
  )
}
