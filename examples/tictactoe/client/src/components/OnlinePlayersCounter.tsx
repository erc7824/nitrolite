import { Users } from "lucide-react";
import { Badge } from "./ui/badge";
import { cn } from "../lib/utils";
import { useEffect } from "react";

interface OnlinePlayersCounterProps {
    className?: string;
    count?: number;
}

export function OnlinePlayersCounter({ className, count = 1 }: OnlinePlayersCounterProps) {
    // For debugging - log count when it changes
    useEffect(() => {
        console.log("OnlinePlayersCounter rendering with count:", count);
    }, [count]);

    return (
        <div className={cn("flex items-center", className)}>
            <Badge variant="secondary" className="bg-fuchsia-950/30 hover:bg-fuchsia-950/40 text-fuchsia-300 border-fuchsia-800/30">
                <Users className="h-3 w-3 mr-1.5" />
                <span>{count} online</span>
            </Badge>
        </div>
    );
}
